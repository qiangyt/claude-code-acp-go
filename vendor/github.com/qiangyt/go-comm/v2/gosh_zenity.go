package comm

import (
	"context"
	"fmt"

	"github.com/ncruces/zenity"
	"mvdan.cc/sh/v3/interp"
)

// ExecZenityHandler zenity 命令处理器
func ExecZenityHandler(ctx context.Context, hc interp.HandlerContext, args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(hc.Stdout, "zenity: missing subcommand")
		return nil
	}

	subCmd := args[0]
	restArgs := args[1:]
	if len(restArgs) == 0 {
		restArgs = []string{}
	}

	switch subCmd {
	case "--error":
		return execZenityError(ctx, hc, restArgs)
	case "--info":
		return execZenityInfo(ctx, hc, restArgs)
	case "--warning":
		return execZenityWarning(ctx, hc, restArgs)
	case "--question":
		return execZenityQuestion(ctx, hc, restArgs)
	default:
		fmt.Fprintln(hc.Stdout, "zenity: unknown subcommand:", subCmd)
		return nil
	}
}

func execZenityError(ctx context.Context, hc interp.HandlerContext, args []string) error {
	text := ""
	if len(args) > 0 {
		text = args[0]
	}
	return zenity.Error(text)
}

func execZenityInfo(ctx context.Context, hc interp.HandlerContext, args []string) error {
	text := ""
	if len(args) > 0 {
		text = args[0]
	}
	return zenity.Info(text)
}

func execZenityWarning(ctx context.Context, hc interp.HandlerContext, args []string) error {
	text := ""
	if len(args) > 0 {
		text = args[0]
	}
	return zenity.Warning(text)
}

func execZenityQuestion(ctx context.Context, hc interp.HandlerContext, args []string) error {
	text := ""
	if len(args) > 0 {
		text = args[0]
	}
	return zenity.Question(text)
}
