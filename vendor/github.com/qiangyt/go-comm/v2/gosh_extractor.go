package comm

import (
	"bytes"
	"strings"

	"github.com/pkg/errors"
	"mvdan.cc/sh/v3/syntax"
)

// ============================================================
// CommandSource 命令来源类型
// ============================================================

// CommandSource 表示命令的来源类型
type CommandSource int

const (
	// SourceDirect 直接调用: echo hello
	SourceDirect CommandSource = iota
	// SourceCmdSubst 命令替换: $(cmd) 或 `cmd`
	SourceCmdSubst
	// SourceProcSubst 进程替换: <(cmd) 或 >(cmd)
	SourceProcSubst
	// SourceSubshell 子 shell: (cmd)
	SourceSubshell
)

// String 返回命令来源的字符串表示
func (s CommandSource) String() string {
	switch s {
	case SourceDirect:
		return "direct"
	case SourceCmdSubst:
		return "command_substitution"
	case SourceProcSubst:
		return "process_substitution"
	case SourceSubshell:
		return "subshell"
	default:
		return "unknown"
	}
}

// ============================================================
// ExtractedCommand 提取的命令信息
// ============================================================

// ExtractedCommandT 从 AST 提取的命令信息
type ExtractedCommandT struct {
	// Name 命令名
	Name string
	// Args 参数列表
	Args []string
	// Source 命令来源类型
	Source CommandSource
	// Position 在源码中的位置
	Position syntax.Pos
	// Envs 环境变量设置 (如 "VAR=value")
	Envs []string
}

// ExtractedCommand 是 ExtractedCommandT 的指针别名
type ExtractedCommand = *ExtractedCommandT

// ============================================================
// CommandExtractor 命令提取器
// ============================================================

// CommandExtractorT 命令提取器，从 shell 命令字符串中提取所有命令
type CommandExtractorT struct {
	parser *syntax.Parser
}

// CommandExtractor 是 CommandExtractorT 的指针别名
type CommandExtractor = *CommandExtractorT

// NewCommandExtractor 创建命令提取器
func NewCommandExtractor() CommandExtractor {
	return &CommandExtractorT{
		parser: syntax.NewParser(),
	}
}

// Extract 从命令字符串中提取所有命令
func (me CommandExtractor) Extract(cmd string) ([]ExtractedCommand, error) {
	if strings.TrimSpace(cmd) == "" {
		return []ExtractedCommand{}, nil
	}

	// 解析命令
	file, err := me.parser.Parse(strings.NewReader(cmd), "")
	if err != nil {
		return nil, errors.Wrapf(err, "parse command: %s", cmd)
	}

	// 提取命令
	var commands []ExtractedCommand
	me.extractFromFile(file, SourceDirect, &commands)

	return commands, nil
}

// extractFromFile 从 File 节点提取命令
func (me CommandExtractor) extractFromFile(file *syntax.File, source CommandSource, commands *[]ExtractedCommand) {
	for _, stmt := range file.Stmts {
		me.extractFromStmt(stmt, source, commands)
	}
}

// extractFromStmt 从 Stmt 节点提取命令
func (me CommandExtractor) extractFromStmt(stmt *syntax.Stmt, source CommandSource, commands *[]ExtractedCommand) {
	if stmt.Cmd == nil {
		return
	}

	switch cmd := stmt.Cmd.(type) {
	case *syntax.CallExpr:
		me.extractFromCallExpr(cmd, source, commands)
	case *syntax.BinaryCmd:
		me.extractFromBinaryCmd(cmd, source, commands)
	case *syntax.Subshell:
		me.extractFromSubshell(cmd, commands)
	case *syntax.IfClause:
		me.extractFromIfClause(cmd, commands)
	case *syntax.WhileClause:
		me.extractFromWhileClause(cmd, commands)
	case *syntax.ForClause:
		me.extractFromForClause(cmd, commands)
	case *syntax.CaseClause:
		me.extractFromCaseClause(cmd, commands)
	case *syntax.Block:
		me.extractFromBlock(cmd, source, commands)
	case *syntax.FuncDecl:
		// 函数声明中的命令
		me.extractFromStmt(cmd.Body, source, commands)
	}
}

// extractFromCallExpr 从 CallExpr 节点提取命令（直接命令调用）
func (me CommandExtractor) extractFromCallExpr(call *syntax.CallExpr, source CommandSource, commands *[]ExtractedCommand) {
	if len(call.Args) == 0 {
		return
	}

	// 提取环境变量
	var envs []string
	args := call.Args

	// 处理内联环境变量 (VAR=value cmd args...)
	if call.Assigns != nil && len(call.Assigns) > 0 {
		for _, assign := range call.Assigns {
			envs = append(envs, assign.Name.Value+"="+me.wordToString(assign.Value))
		}
	}

	// 第一个参数是命令名
	cmdName := me.wordToString(args[0])
	if cmdName == "" {
		return
	}

	// 提取参数
	var cmdArgs []string
	for i := 1; i < len(args); i++ {
		arg := me.wordToString(args[i])
		if arg != "" {
			cmdArgs = append(cmdArgs, arg)
		}
	}

	*commands = append(*commands, &ExtractedCommandT{
		Name:     cmdName,
		Args:     cmdArgs,
		Source:   source,
		Position: call.Pos(),
		Envs:     envs,
	})

	// 检查参数中是否有命令替换
	for _, arg := range args {
		me.extractWordSubcommands(arg, commands)
	}
}

// extractFromBinaryCmd 从 BinaryCmd 节点提取命令（管道和条件命令）
func (me CommandExtractor) extractFromBinaryCmd(binary *syntax.BinaryCmd, source CommandSource, commands *[]ExtractedCommand) {
	// 提取左边的命令
	me.extractFromStmt(binary.X, source, commands)
	// 提取右边的命令
	me.extractFromStmt(binary.Y, source, commands)
}

// extractFromSubshell 从 Subshell 节点提取命令
func (me CommandExtractor) extractFromSubshell(subshell *syntax.Subshell, commands *[]ExtractedCommand) {
	for _, stmt := range subshell.Stmts {
		me.extractFromStmt(stmt, SourceSubshell, commands)
	}
}

// extractFromIfClause 从 IfClause 节点提取命令
func (me CommandExtractor) extractFromIfClause(ifClause *syntax.IfClause, commands *[]ExtractedCommand) {
	// 条件部分
	for _, stmt := range ifClause.Cond {
		me.extractFromStmt(stmt, SourceDirect, commands)
	}
	// then 部分
	for _, stmt := range ifClause.Then {
		me.extractFromStmt(stmt, SourceDirect, commands)
	}
	// else 部分 (Else 是 *IfClause 类型，可以是 elif 或 else)
	if ifClause.Else != nil {
		me.extractFromIfClause(ifClause.Else, commands)
	}
}

// extractFromWhileClause 从 WhileClause 节点提取命令
func (me CommandExtractor) extractFromWhileClause(whileClause *syntax.WhileClause, commands *[]ExtractedCommand) {
	// 条件部分
	for _, stmt := range whileClause.Cond {
		me.extractFromStmt(stmt, SourceDirect, commands)
	}
	// 循环体
	for _, stmt := range whileClause.Do {
		me.extractFromStmt(stmt, SourceDirect, commands)
	}
}

// extractFromForClause 从 ForClause 节点提取命令
func (me CommandExtractor) extractFromForClause(forClause *syntax.ForClause, commands *[]ExtractedCommand) {
	// 循环体
	for _, stmt := range forClause.Do {
		me.extractFromStmt(stmt, SourceDirect, commands)
	}
}

// extractFromCaseClause 从 CaseClause 节点提取命令
func (me CommandExtractor) extractFromCaseClause(caseClause *syntax.CaseClause, commands *[]ExtractedCommand) {
	for _, item := range caseClause.Items {
		for _, stmt := range item.Stmts {
			me.extractFromStmt(stmt, SourceDirect, commands)
		}
	}
}

// extractFromBlock 从 Block 节点提取命令
func (me CommandExtractor) extractFromBlock(block *syntax.Block, source CommandSource, commands *[]ExtractedCommand) {
	for _, stmt := range block.Stmts {
		me.extractFromStmt(stmt, source, commands)
	}
}

// extractWordSubcommands 从 Word 中提取命令替换
func (me CommandExtractor) extractWordSubcommands(word *syntax.Word, commands *[]ExtractedCommand) {
	if word == nil {
		return
	}

	for _, part := range word.Parts {
		switch p := part.(type) {
		case *syntax.CmdSubst:
			// 命令替换 $(cmd) 或 `cmd`
			for _, stmt := range p.Stmts {
				me.extractFromStmt(stmt, SourceCmdSubst, commands)
			}
		case *syntax.DblQuoted:
			// 双引号中可能有命令替换
			for _, qp := range p.Parts {
				if cs, ok := qp.(*syntax.CmdSubst); ok {
					for _, stmt := range cs.Stmts {
						me.extractFromStmt(stmt, SourceCmdSubst, commands)
					}
				}
			}
		case *syntax.ParamExp:
			// 参数展开中可能有命令替换 ${...}
			// Exp.Word 可能包含命令替换
			if p.Exp != nil && p.Exp.Word != nil {
				me.extractWordSubcommands(p.Exp.Word, commands)
			}
		case *syntax.ProcSubst:
			// 进程替换 <(cmd) 或 >(cmd)
			for _, stmt := range p.Stmts {
				me.extractFromStmt(stmt, SourceProcSubst, commands)
			}
		}
	}
}

// wordToString 将 Word 转换为字符串（保留变量、引号等）
func (me CommandExtractor) wordToString(word *syntax.Word) string {
	if word == nil {
		return ""
	}

	// 首先尝试 Lit()，如果是纯字面量就直接返回
	if lit := word.Lit(); lit != "" {
		return lit
	}

	// 否则使用 Printer 打印完整的字符串表示
	var buf bytes.Buffer
	printer := syntax.NewPrinter()
	if err := printer.Print(&buf, word); err != nil {
		return ""
	}
	return buf.String()
}
