package edit

import (
	"testing"

	plan "github.com/ryancragun/terraform-plan-editor/internal/proto/v1"
	"github.com/stretchr/testify/require"
)

func requireEqualResourceInstanceChanges(t *testing.T, e, a []*plan.ResourceInstanceChange) {
	t.Helper()
	req := require.New(t)

	if e == nil {
		req.Nil(a)
	} else {
		req.NotNil(a)
		req.Len(a, len(e))
		for i, erc := range e {
			req.Equal(erc.Addr, a[i].Addr)
			req.Equal(erc.Provider, a[i].Provider)
			requireEqualChange(t, erc.Change, a[i].Change)
		}
	}
}

func requireEqualChange(t *testing.T, e, a *plan.Change) {
	t.Helper()
	req := require.New(t)

	if e == nil {
		req.Nil(a)
	} else {
		req.NotNil(a)
		req.Equal(e.Action, a.Action)
		req.Len(a.BeforeSensitivePaths, len(a.BeforeSensitivePaths))
		for i, eps := range e.BeforeSensitivePaths {
			requireEqualPath(t, eps, a.BeforeSensitivePaths[i])
		}
	}
}

func requireEqualPath(t *testing.T, e, a *plan.Path) {
	t.Helper()
	req := require.New(t)

	if e == nil {
		req.Nil(a)
	} else {
		req.NotNil(a)
		req.Len(a.Steps, len(a.Steps))
		for i, es := range e.Steps {
			if es.Selector == nil {
				req.Nil(a.Steps[i].Selector)
			} else {
				req.NotNil(a.Steps[i].Selector)
				req.Equal(es.GetAttributeName(), a.Steps[i].GetAttributeName())
				req.Equal(es.GetElementKey(), a.Steps[i].GetElementKey())
			}
		}
	}
}

func requireEqualPlan(t *testing.T, e, a *plan.Plan) {
	req := require.New(t)

	req.Equal(e.Version, a.Version)
	req.Equal(e.Applyable, a.Applyable)
	req.Equal(e.Complete, a.Complete)
	req.Equal(e.TerraformVersion, a.TerraformVersion)
	req.Equal(e.Timestamp, a.Timestamp)
	if e.Backend == nil {
		req.Nil(a.Backend)
	} else {
		req.NotNil(a.Backend)
		req.Equal(e.Backend.Type, a.Backend.Type)
		req.Equal(e.Backend.Workspace, a.Backend.Workspace)
	}
	requireEqualResourceInstanceChanges(t, e.ResourceChanges, a.ResourceChanges)
	requireEqualResourceInstanceChanges(t, e.ResourceDrift, a.ResourceDrift)
}

func TestCombinePlan(t *testing.T) {
	t.Parallel()

	sans := &plan.Plan{
		Version:          3,
		Applyable:        true,
		Complete:         true,
		TerraformVersion: "1.9.1",
		Backend: &plan.Backend{
			Type:      "local",
			Workspace: "default",
		},
		Timestamp: "2024-07-08T17:19:28Z",
		ResourceChanges: []*plan.ResourceInstanceChange{
			{
				Addr:     "enos_bundle_install.vault",
				Provider: `provider[\"registry.terraform.io/hashicorp-forge/enos\"]`,
				Change: &plan.Change{
					Action: plan.Action_UPDATE,
					BeforeSensitivePaths: []*plan.Path{
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "token"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "url"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "url"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "username"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "sha256"},
								},
							},
						},
					},
					AfterSensitivePaths: []*plan.Path{
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
							},
						},
					},
				},
			},
		},
		ResourceDrift: []*plan.ResourceInstanceChange{
			{
				Addr:     "enos_bundle_install.vault",
				Provider: `provider[\"registry.terraform.io/hashicorp-forge/enos\"]`,
				Change: &plan.Change{
					Action: plan.Action_UPDATE,
					BeforeSensitivePaths: []*plan.Path{
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
							},
						},
					},
					AfterSensitivePaths: []*plan.Path{
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "token"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "url"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "url"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "username"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "sha256"},
								},
							},
						},
					},
				},
			},
		},
	}

	only := &plan.Plan{
		ResourceChanges: []*plan.ResourceInstanceChange{
			{
				Addr:     "enos_bundle_install.vault",
				Provider: `provider[\"registry.terraform.io/hashicorp-forge/enos\"]`,
				Change: &plan.Change{
					Action: plan.Action_UPDATE,
					Values: []*plan.DynamicValue{
						{
							Msgpack: []byte(`\x86\xabartifactory\x84\xa6sha256\xd9@d01a82111133908167a5a140604ab3ec8fd18601758376a5f8e9dd54c7703373\xa5token\xd9Isnip\xa3url\xda\x01\x13http://artifactory.com/7fb88d4d3d0a36ffc78a522d870492e5791bae1b0640232ce4c6d69cc22cf520/store/f45845666b4e552bfc8ca775834a3ef6fc097fe0-1a2809da73e5896b6f766b395ff6e1804f876c45.zip\xa8username\xb5name@example.com\xabdestination\xae/opt/bin/vault\xa2id\xa6static\xa4path\xc0\xa7release\xc0\xa9transport\x92\xc4[[\"object\",{\"ssh\":[\"object\",{\"host\":\"string\",\"private_key_path\":\"string\",\"user\":\"string\"}]}]\x81\xa3ssh\x83\xa4host\xae0.0.0.0\xb0private_key_path\xd9O/ssh.pem\xa4user\xa6ubuntu`),
						},
						{
							Msgpack: []byte(`\x86\xabartifactory\x84\xa6sha256\xd9@d01a82111133908167a5a140604ab3ec8fd18601758376a5f8e9dd54c7703373\xa5token\xd9Isnip\xa3url\xda\x01\x13http://artifactory.com/7fb88d4d3d0a36ffc78a522d870492e5791bae1b0640232ce4c6d69cc22cf520/store/f45845666b4e552bfc8ca775834a3ef6fc097fe0-1a2809da73e5896b6f766b395ff6e1804f876c45.zip\xa8username\xb5name@example.com\xabdestination\xae/opt/bin/vault\xa2id\xa6static\xa4path\xc0\xa7release\xc0\xa9transport\x92\xc4[[\"object\",{\"ssh\":[\"object\",{\"host\":\"string\",\"private_key_path\":\"string\",\"user\":\"string\"}]}]\x81\xa3ssh\x83\xa4host\xae0.0.0.0\xb0private_key_path\xd9O/ssh.pem\xa4user\xa6ubuntu`),
						},
					},
				},
			},
		},
		ResourceDrift: []*plan.ResourceInstanceChange{
			{
				Addr:     "enos_bundle_install.vault",
				Provider: `provider[\"registry.terraform.io/hashicorp-forge/enos\"]`,
				Change: &plan.Change{
					Action: plan.Action_UPDATE,
					Values: []*plan.DynamicValue{
						{
							Msgpack: []byte(`\x86\xabartifactory\x84\xa6sha256\xd9@d01a82111133908167a5a140604ab3ec8fd18601758376a5f8e9dd54c7703373\xa5token\xd9Isnip\xa3url\xda\x01\x13http://artifactory.com/7fb88d4d3d0a36ffc78a522d870492e5791bae1b0640232ce4c6d69cc22cf520/store/f45845666b4e552bfc8ca775834a3ef6fc097fe0-1a2809da73e5896b6f766b395ff6e1804f876c45.zip\xa8username\xb5name@example.com\xabdestination\xae/opt/bin/vault\xa2id\xa6static\xa4path\xc0\xa7release\xc0\xa9transport\x92\xc4[[\"object\",{\"ssh\":[\"object\",{\"host\":\"string\",\"private_key_path\":\"string\",\"user\":\"string\"}]}]\x81\xa3ssh\x83\xa4host\xae0.0.0.0\xb0private_key_path\xd9O/ssh.pem\xa4user\xa6ubuntu`),
						},
						{
							Msgpack: []byte(`\x86\xabartifactory\x84\xa6sha256\xd9@d01a82111133908167a5a140604ab3ec8fd18601758376a5f8e9dd54c7703373\xa5token\xd9Isnip\xa3url\xda\x01\x13http://artifactory.com/7fb88d4d3d0a36ffc78a522d870492e5791bae1b0640232ce4c6d69cc22cf520/store/f45845666b4e552bfc8ca775834a3ef6fc097fe0-1a2809da73e5896b6f766b395ff6e1804f876c45.zip\xa8username\xb5name@example.com\xabdestination\xae/opt/bin/vault\xa2id\xa6static\xa4path\xc0\xa7release\xc0\xa9transport\x92\xc4[[\"object\",{\"ssh\":[\"object\",{\"host\":\"string\",\"private_key_path\":\"string\",\"user\":\"string\"}]}]\x81\xa3ssh\x83\xa4host\xae0.0.0.0\xb0private_key_path\xd9O/ssh.pem\xa4user\xa6ubuntu`),
						},
					},
				},
			},
		},
	}

	expected := &plan.Plan{
		Version:          3,
		Applyable:        true,
		Complete:         true,
		TerraformVersion: "1.9.1",
		Backend: &plan.Backend{
			Type:      "local",
			Workspace: "default",
		},
		Timestamp: "2024-07-08T17:19:28Z",
		ResourceChanges: []*plan.ResourceInstanceChange{
			{
				Addr:     "enos_bundle_install.vault",
				Provider: `provider[\"registry.terraform.io/hashicorp-forge/enos\"]`,
				Change: &plan.Change{
					Action: plan.Action_UPDATE,
					Values: []*plan.DynamicValue{
						{
							Msgpack: []byte(`\x86\xabartifactory\x84\xa6sha256\xd9@d01a82111133908167a5a140604ab3ec8fd18601758376a5f8e9dd54c7703373\xa5token\xd9Isnip\xa3url\xda\x01\x13http://artifactory.com/7fb88d4d3d0a36ffc78a522d870492e5791bae1b0640232ce4c6d69cc22cf520/store/f45845666b4e552bfc8ca775834a3ef6fc097fe0-1a2809da73e5896b6f766b395ff6e1804f876c45.zip\xa8username\xb5name@example.com\xabdestination\xae/opt/bin/vault\xa2id\xa6static\xa4path\xc0\xa7release\xc0\xa9transport\x92\xc4[[\"object\",{\"ssh\":[\"object\",{\"host\":\"string\",\"private_key_path\":\"string\",\"user\":\"string\"}]}]\x81\xa3ssh\x83\xa4host\xae0.0.0.0\xb0private_key_path\xd9O/ssh.pem\xa4user\xa6ubuntu`),
						},
						{
							Msgpack: []byte(`\x86\xabartifactory\x84\xa6sha256\xd9@d01a82111133908167a5a140604ab3ec8fd18601758376a5f8e9dd54c7703373\xa5token\xd9Isnip\xa3url\xda\x01\x13http://artifactory.com/7fb88d4d3d0a36ffc78a522d870492e5791bae1b0640232ce4c6d69cc22cf520/store/f45845666b4e552bfc8ca775834a3ef6fc097fe0-1a2809da73e5896b6f766b395ff6e1804f876c45.zip\xa8username\xb5name@example.com\xabdestination\xae/opt/bin/vault\xa2id\xa6static\xa4path\xc0\xa7release\xc0\xa9transport\x92\xc4[[\"object\",{\"ssh\":[\"object\",{\"host\":\"string\",\"private_key_path\":\"string\",\"user\":\"string\"}]}]\x81\xa3ssh\x83\xa4host\xae0.0.0.0\xb0private_key_path\xd9O/ssh.pem\xa4user\xa6ubuntu`),
						},
					},
					BeforeSensitivePaths: []*plan.Path{
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "token"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "url"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "url"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "username"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "sha256"},
								},
							},
						},
					},
					AfterSensitivePaths: []*plan.Path{
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
							},
						},
					},
				},
			},
		},
		ResourceDrift: []*plan.ResourceInstanceChange{
			{
				Addr:     "enos_bundle_install.vault",
				Provider: `provider[\"registry.terraform.io/hashicorp-forge/enos\"]`,
				Change: &plan.Change{
					Action: plan.Action_UPDATE,
					Values: []*plan.DynamicValue{
						{
							Msgpack: []byte(`\x86\xabartifactory\x84\xa6sha256\xd9@d01a82111133908167a5a140604ab3ec8fd18601758376a5f8e9dd54c7703373\xa5token\xd9Isnip\xa3url\xda\x01\x13http://artifactory.com/7fb88d4d3d0a36ffc78a522d870492e5791bae1b0640232ce4c6d69cc22cf520/store/f45845666b4e552bfc8ca775834a3ef6fc097fe0-1a2809da73e5896b6f766b395ff6e1804f876c45.zip\xa8username\xb5name@example.com\xabdestination\xae/opt/bin/vault\xa2id\xa6static\xa4path\xc0\xa7release\xc0\xa9transport\x92\xc4[[\"object\",{\"ssh\":[\"object\",{\"host\":\"string\",\"private_key_path\":\"string\",\"user\":\"string\"}]}]\x81\xa3ssh\x83\xa4host\xae0.0.0.0\xb0private_key_path\xd9O/ssh.pem\xa4user\xa6ubuntu`),
						},
						{
							Msgpack: []byte(`\x86\xabartifactory\x84\xa6sha256\xd9@d01a82111133908167a5a140604ab3ec8fd18601758376a5f8e9dd54c7703373\xa5token\xd9Isnip\xa3url\xda\x01\x13http://artifactory.com/7fb88d4d3d0a36ffc78a522d870492e5791bae1b0640232ce4c6d69cc22cf520/store/f45845666b4e552bfc8ca775834a3ef6fc097fe0-1a2809da73e5896b6f766b395ff6e1804f876c45.zip\xa8username\xb5name@example.com\xabdestination\xae/opt/bin/vault\xa2id\xa6static\xa4path\xc0\xa7release\xc0\xa9transport\x92\xc4[[\"object\",{\"ssh\":[\"object\",{\"host\":\"string\",\"private_key_path\":\"string\",\"user\":\"string\"}]}]\x81\xa3ssh\x83\xa4host\xae0.0.0.0\xb0private_key_path\xd9O/ssh.pem\xa4user\xa6ubuntu`),
						},
					},
					BeforeSensitivePaths: []*plan.Path{
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
							},
						},
					},
					AfterSensitivePaths: []*plan.Path{
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "token"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "url"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "url"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "username"},
								},
							},
						},
						{
							Steps: []*plan.Path_Step{
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "artifactory"},
								},
								{
									Selector: &plan.Path_Step_AttributeName{AttributeName: "sha256"},
								},
							},
						},
					},
				},
			},
		},
	}

	comb, err := combinePlans(sans, only)
	require.NoError(t, err)
	requireEqualPlan(t, expected, comb)
}
