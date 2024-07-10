package edit

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ugorji/go/codec"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	ctymsgpack "github.com/zclconf/go-cty/cty/msgpack"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"

	plan "github.com/ryancragun/terraform-plan-editor/internal/proto/v1"
)

type Config struct {
	TextEditorCmd string
	BinEditorCmd  string
	PlanPath      string
	DstPath       string
}

type Editor struct {
	*Config
}

func New(cfg *Config) *Editor {
	return &Editor{Config: cfg}
}

func (e *Editor) Edit() error {
	dir, err := e.unzipPlan()
	if err != nil {
		return err
	}

	if err = e.editFilesIn(dir); err != nil {
		return err
	}

	if err = e.zipPlan(dir); err != nil {
		return err
	}

	return nil
}

func (e *Editor) unzipPlan() (string, error) {
	if e == nil || e.PlanPath == "" {
		return "", errors.New("you must provide a path to a Terraform plan")
	}
	path := e.PlanPath

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("unable to locate Terraform plan: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "terraform-plan-edit")
	if err != nil {
		return "", err
	}

	reader, err := zip.OpenReader(abs)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	for _, zipFile := range reader.File {
		zf, err := zipFile.Open()
		if err != nil {
			return "", err
		}
		defer zf.Close()

		p := filepath.Join(tmpDir, zipFile.Name)
		if err = os.MkdirAll(filepath.Dir(p), 0o770); err != nil {
			return "", err
		}
		f, err := os.Create(p)
		if err != nil {
			return "", err
		}
		defer f.Close()

		fmt.Println("inflate: " + p)
		_, err = io.Copy(f, zf)
		if err != nil {
			return "", err
		}
	}

	return tmpDir, nil
}

func (d *Editor) zipPlan(dir string) error {
	if dir == "" {
		return errors.New("zip plan: no directory given")
	}

	newPlan, err := os.Create(d.DstPath)
	if err != nil {
		return err
	}
	defer func() {
		fmt.Println("write: " + d.DstPath)
	}()

	defer newPlan.Close()

	archive := zip.NewWriter(newPlan)
	defer archive.Close()

	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		// Make sure we keep the correct directory structure
		name, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		fmt.Println("deflate: " + path)
		w, err := archive.CreateHeader(&zip.FileHeader{
			Name:     name,
			Method:   zip.Deflate,
			Modified: time.Now(),
		})
		if err != nil {
			return err
		}

		bytes, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		written, err := w.Write(bytes)
		if err != nil {
			return err
		}

		if len(bytes) != written {
			return fmt.Errorf("expected %d bytes to be written, got %d", len(bytes), written)
		}

		return nil
	})
}

func copy(in proto.Message, out proto.Message) error {
	if in == nil || out == nil {
		return errors.New("you must provide a source and destination message to copy")
	}

	bytes, err := prototext.Marshal(in)
	if err != nil {
		return err
	}

	if len(bytes) == 0 {
		return nil
	}

	return prototext.Unmarshal(bytes, out)
}

func editTFPlan(config *Config, path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if len(bytes) == 0 {
		return nil
	}

	origPlan := &plan.Plan{}
	err = proto.Unmarshal(bytes, origPlan)
	if err != nil {
		return err
	}

	// We edit the plan in two stages: once for everything that is not msgpack/dynamic values and
	// then for every value that is. We then combine our values together by applying the msgpack only
	// plan on top. This allows for more specific editing of values that are msgpack encoded and
	// otherwise very difficult to edit as text.
	planSansMsgpack, err := editTFPlanNoMsgPack(path, config.TextEditorCmd, origPlan)
	if err != nil {
		return err
	}

	planOnlyMsgpack, err := editTFPlanOnlyMsgPack(path, config, origPlan)
	if err != nil {
		return err
	}

	np, err := combinePlans(planSansMsgpack, planOnlyMsgpack)
	if err != nil {
		return err
	}

	npBytes, err := proto.Marshal(np)
	if err != nil {
		return err
	}

	if err = file.Truncate(0); err != nil {
		return err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = file.Write(npBytes)
	return err
}

func combinePlans(sans *plan.Plan, only *plan.Plan) (*plan.Plan, error) {
	np := &plan.Plan{}
	if err := copy(sans, np); err != nil {
		return nil, err
	}

	for k, v := range only.GetVariables() {
		if np.GetVariables() == nil {
			np.Variables = make(map[string]*plan.DynamicValue)
		}

		np.Variables[k] = v
	}

	for ic, c := range only.GetResourceChanges() {
		if np.GetResourceChanges() == nil {
			np.ResourceChanges = []*plan.ResourceInstanceChange{}
		}

		if v := c.GetChange().GetValues(); v != nil {
			if np.GetResourceChanges()[ic] == nil {
				np.ResourceChanges[ic] = &plan.ResourceInstanceChange{}
			}
			if np.GetResourceChanges()[ic].GetChange() == nil {
				np.ResourceChanges[ic].Change = &plan.Change{}
			}

			np.ResourceChanges[ic].Change.Values = v
		}
	}

	for id, c := range only.GetResourceDrift() {
		if np.GetResourceDrift() == nil {
			np.ResourceDrift = []*plan.ResourceInstanceChange{}
		}

		if v := c.GetChange().GetValues(); v != nil {
			if np.GetResourceDrift()[id] == nil {
				np.ResourceDrift[id] = &plan.ResourceInstanceChange{}
			}
			if np.GetResourceDrift()[id].GetChange() == nil {
				np.ResourceDrift[id].Change = &plan.Change{}
			}

			np.ResourceDrift[id].Change.Values = v
		}
	}

	for ic, d := range only.GetDeferredChanges() {
		if np.GetDeferredChanges() == nil {
			np.DeferredChanges = []*plan.DeferredResourceInstanceChange{}
		}

		if np.GetDeferredChanges()[ic] == nil {
			np.DeferredChanges[ic] = &plan.DeferredResourceInstanceChange{}
		}

		if np.GetDeferredChanges()[ic].GetChange() == nil {
			np.DeferredChanges[ic].Change = &plan.ResourceInstanceChange{}
		}

		if cc := np.DeferredChanges[ic].GetChange().GetChange(); cc == nil {
			np.DeferredChanges[ic].Change.Change = &plan.Change{}
		}

		if values := d.GetChange().GetChange().GetValues(); values != nil {
			np.DeferredChanges[ic].Change.Change.Values = values
		}
	}

	for io, o := range only.GetOutputChanges() {
		if np.GetOutputChanges() == nil {
			np.OutputChanges = []*plan.OutputChange{}
		}

		if values := o.GetChange().GetValues(); values != nil {
			if oc := np.OutputChanges[io]; oc == nil {
				np.OutputChanges[io] = &plan.OutputChange{}
			}

			if c := np.OutputChanges[io].GetChange(); c == nil {
				np.OutputChanges[io].Change = &plan.Change{}
			}

			np.OutputChanges[io].Change.Values = values
		}
	}

	if c := only.GetBackend().GetConfig(); c != nil {
		if np.GetBackend() == nil {
			np.Backend = &plan.Backend{}
		}

		np.Backend.Config = c
	}

	return np, nil
}

func editTFPlanNoMsgPack(
	path string,
	editorCmd string,
	p *plan.Plan,
) (*plan.Plan, error) {
	np := &plan.Plan{}
	if err := copy(p, np); err != nil {
		return nil, err
	}

	// Remove most of the msgpack values and covert the plan to JSON to allow editing it. We
	// handle each msgpack/DynamicValue in a separate pass since they have to be decoded.
	if v := np.GetVariables(); v != nil {
		np.Variables = nil
	}

	for ic, c := range np.GetResourceChanges() {
		if c.GetChange().GetValues() != nil {
			np.ResourceChanges[ic].Change.Values = nil
		}
	}

	for id, d := range np.GetResourceDrift() {
		if d.GetChange().GetValues() != nil {
			np.ResourceDrift[id].Change.Values = nil
		}
	}

	for id, d := range np.GetDeferredChanges() {
		if change := d.GetChange().GetChange(); change != nil {
			np.DeferredChanges[id].Change.Change.Values = nil
		}
	}

	for io, o := range np.GetOutputChanges() {
		if o.GetChange().GetValues() != nil {
			np.OutputChanges[io].Change.Values = nil
		}
	}

	if c := np.GetBackend().GetConfig(); c != nil {
		np.Backend.Config = nil
	}

	protoJSON := protojson.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}
	jsonBytes, err := protoJSON.Marshal(np)
	if err != nil {
		return nil, err
	}

	tmpFilePath := filepath.Join(filepath.Dir(path), "tfplan-sans-dynamic-values.json")
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFilePath)

	_, err = tmpFile.Write(jsonBytes)
	if err != nil {
		return nil, err
	}
	tmpFile.Close()

	err = editFile(editorCmd, tmpFilePath)
	if err != nil {
		return nil, err
	}

	tmpFile, err = os.Open(tmpFilePath)
	if err != nil {
		return nil, err
	}
	defer tmpFile.Close()

	bytes, err := io.ReadAll(tmpFile)
	if err != nil {
		return nil, err
	}

	// Write the plan in the binary format
	editedP := &plan.Plan{}
	err = protojson.Unmarshal(bytes, editedP)
	return editedP, err
}

func editTFPlanOnlyMsgPack(path string, config *Config, p *plan.Plan) (*plan.Plan, error) {
	np := &plan.Plan{}
	if err := copy(p, np); err != nil {
		return nil, err
	}

	// Update most of the msgpack values by converting them into JSON to allow easier editing.
	for k, v := range np.GetVariables() {
		if v.GetMsgpack() == nil {
			continue
		}

		if err := editDynamicValue(path, config, v, "plan_variable_"+k); err != nil {
			return nil, err
		}
	}

	for _, c := range np.GetResourceChanges() {
		for iv, v := range c.GetChange().GetValues() {
			if v.GetMsgpack() == nil {
				continue
			}

			if err := editDynamicValue(path, config, v, fmt.Sprintf("resource_change_%s_%d", c.GetAddr(), iv)); err != nil {
				return nil, err
			}
		}
	}

	for _, d := range np.GetResourceDrift() {
		for iv, v := range d.GetChange().GetValues() {
			if v.GetMsgpack() == nil {
				continue
			}

			if err := editDynamicValue(path, config, v, fmt.Sprintf("resource_drift_%s_%d", d.GetAddr(), iv)); err != nil {
				return nil, err
			}
		}
	}

	for _, d := range np.GetDeferredChanges() {
		for iv, v := range d.GetChange().GetChange().GetValues() {
			if v.GetMsgpack() == nil {
				continue
			}

			if err := editDynamicValue(path, config, v, fmt.Sprintf("deferred_change_%s_%d", d.GetChange().GetAddr(), iv)); err != nil {
				return nil, err
			}
		}
	}

	for _, d := range np.GetOutputChanges() {
		for iv, v := range d.GetChange().GetValues() {
			if v.GetMsgpack() == nil {
				continue
			}

			if err := editDynamicValue(path, config, v, fmt.Sprintf("output_change_%s_%d", d.Name, iv)); err != nil {
				return nil, err
			}
		}
	}

	if c := np.GetBackend().GetConfig(); c != nil {
		if err := editDynamicValue(path, config, c, "backend_config"); err != nil {
			return nil, err
		}
	}

	return np, nil
}

func editDynamicValue(path string, config *Config, d *plan.DynamicValue, desc string) error {
	if d == nil {
		return nil
	}

	bytes := d.GetMsgpack()
	if len(bytes) == 0 {
		return nil
	}

	// If we're editing values that have been encoded with as a cty.DynamicPseudoType this will fail
	// and we'll fall back on our raw strategy.
	var err error
	d.Msgpack, err = editDynamicValueKnownCTYType(path, config.TextEditorCmd, bytes, desc)
	if err == nil {
		return nil
	}

	d.Msgpack, err = editDynamicValueUnknownType(path, config.BinEditorCmd, bytes, desc)
	return err
}

func editDynamicValueUnknownType(path string, editorCmd string, bytes []byte, desc string) ([]byte, error) {
	if len(bytes) == 0 {
		return bytes, nil
	}

	msgpackHandle := &codec.MsgpackHandle{}
	msgpackHandle.WriteExt = true

	dec := codec.NewDecoderBytes(bytes, msgpackHandle)
	v := map[any]any{}
	err := dec.Decode(v)
	if err != nil {
		return nil, fmt.Errorf("cannot edit dynamic value: unable to decode Msgpack data for editing: %w", err)
	}

	msgpackBytes := []byte{}
	enc := codec.NewEncoderBytes(&msgpackBytes, msgpackHandle)
	err = enc.Encode(v)
	if err != nil {
		return nil, fmt.Errorf("cannot edit dynamic value: unable to encode data type to JSON for editing: %w", err)
	}

	tmpFilePath := filepath.Join(filepath.Dir(path), "dynamic-value-unknown-type-"+strings.ReplaceAll(desc, " ", "-"))
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFilePath)

	_, err = tmpFile.Write(msgpackBytes)
	if err != nil {
		return nil, err
	}
	tmpFile.Close()

	err = editFile(editorCmd, tmpFilePath)
	if err != nil {
		return nil, err
	}

	tmpFile, err = os.Open(tmpFilePath)
	if err != nil {
		return nil, err
	}
	defer tmpFile.Close()

	return io.ReadAll(tmpFile)
}

func editDynamicValueKnownCTYType(path string, editorCmd string, bytes []byte, desc string) ([]byte, error) {
	if len(bytes) == 0 {
		return bytes, nil
	}

	typ, err := ctymsgpack.ImpliedType(bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot edit dynamic value: %s, unable to infer data type: %w", desc, err)
	}

	val, err := ctymsgpack.Unmarshal(bytes, typ)
	if err != nil {
		return nil, fmt.Errorf("cannot edit dynamic value: %s, unable to decode data type: %w", typ, err)
	}

	jsonBytes, err := ctyjson.Marshal(val, typ)
	if err != nil {
		return nil, fmt.Errorf("cannot edit dynamic value: %s, unable to encode data type to JSON for editing: %w", typ, err)
	}

	tmpFilePath := filepath.Join(filepath.Dir(path), "dynamic-value-known-type-"+strings.ReplaceAll(desc, " ", "-"))
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFilePath)

	_, err = tmpFile.Write(jsonBytes)
	if err != nil {
		return nil, err
	}
	tmpFile.Close()

	err = editFile(editorCmd, tmpFilePath)
	if err != nil {
		return nil, err
	}

	tmpFile, err = os.Open(tmpFilePath)
	if err != nil {
		return nil, err
	}
	defer tmpFile.Close()

	jsonBytes, err = io.ReadAll(tmpFile)
	if err != nil {
		return nil, err
	}

	val, err = ctyjson.Unmarshal(jsonBytes, typ)
	if err != nil {
		return nil, fmt.Errorf("failed to encode edited dynamic value: %w", err)
	}

	bytes, err = ctymsgpack.Marshal(val, typ)
	if err != nil {
		return nil, fmt.Errorf("failed to encode edited dynamic value: %w", err)
	}

	return bytes, nil
}

func editFile(editorCmd string, path string) error {
	args := append(strings.Split(editorCmd, " "), path)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	fmt.Println("edit: " + cmd.String())
	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}

func (e *Editor) editFilesIn(dir string) error {
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, "/tfplan") {
			return editTFPlan(e.Config, path)
		} else {
			return editFile(e.Config.TextEditorCmd, path)
		}
	})
}
