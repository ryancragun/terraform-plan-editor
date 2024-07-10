![Test](https://github.com/ryancragun/terraform-plan-editor/actions/workflows/test.yml/badge.svg)

# terraform-plan-editor

This is a hacky little tool that I wrote to sanitize secrets out of terraform plans. It is very
rudimentary and using it is a bit like performing surgery with a spoon.

> [!CAUTION]
> Under no circumstances should you use this utility to edit plans for real infrastructure

The tool unpacks an existing Terrform plan created with `terraform plan -out=tf.plan` and allow
you to make edits to every file in the plan. The plan contents themselves, especially the binary
`tfplan`, is fairly complicated to edit. As such, you'll likely need to specify both a text editor
and binary editor if your plan includes resources that utilize the DynamicPseudoType. Where possible
we try to render binary data as JSON when editing. Some msgpack binary data connot be easily round
tripped without type information. In those cases it will have you edit the binary chunk with a
binary editor.

## Usage

```shell
go run ./ --editor=nvim --bin-editor='nvim -b' ../path/to/source/tf.plan ./path/to/edited.plan
```

> [!NOTE]
> I have only tested this with nvim as both the text editor and binary editor.
