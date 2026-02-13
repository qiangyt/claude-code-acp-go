module claude-code-acp-go

go 1.24

replace github.com/qiangyt/go-comm/v2 => /data1/baton/go-comm

replace github.com/coder/acp-go-sdk => /data1/baton/refer/coder/acp-go-sdk

require (
	github.com/coder/acp-go-sdk v0.0.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
