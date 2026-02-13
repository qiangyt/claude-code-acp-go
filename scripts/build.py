#!/usr/bin/env python3
"""跨平台构建脚本 - 仅使用 Python 标准库"""

import os
import subprocess
import sys
from pathlib import Path
from typing import Callable, Literal

# 项目常量
PROJECT_NAME = "claude-code-acp-go"


def run(cmd: str, check: bool = True) -> int:
    """执行命令并返回退出码"""
    result = subprocess.run(cmd, shell=True, check=check)
    return result.returncode


def run_quiet(cmd: str) -> int:
    """静默执行命令，不打印输出"""
    return subprocess.run(cmd, shell=True, capture_output=True).returncode


def docker_container_running(container_name: str) -> bool:
    """检查 Docker 容器是否运行中"""
    return run_quiet(f"docker ps -q -f name={container_name}") == 0


def task_mod() -> None:
    """下载并整理 Go 模块"""
    run("go mod download")
    run("go mod tidy")
    run("go mod vendor")


def task_check() -> None:
    """格式化代码并运行检查"""
    run("gofumpt -l -w .")
    #run("staticcheck ./...")
    #run("golangci-lint run")


def task_test() -> None:
    """运行测试并生成覆盖率报告（忽略 legacy 包）"""
    # 创建 build 目录
    Path("build").mkdir(exist_ok=True)
    run("go test ./... -timeout 60s -covermode=count -coverprofile=build/coverage.out")
    run("go tool cover -html=./build/coverage.out -o ./build/coverage.html")


def task_build() -> None:
    """构建当前平台"""
    # 创建 build 目录
    Path("build").mkdir(exist_ok=True)
    run("go build -trimpath -ldflags=\"-s -w\" -o ./build/ ./cmd/...")


def task_build_all() -> None:
    """构建所有平台"""
    run("goreleaser build --clean --snapshot")


def task_release() -> None:
    """构建并发布"""
    run("goreleaser release --clean")


def task_init() -> None:
    """初始化开发环境"""
    tools = [
        ("github.com/golang/mock/mockgen", "v1.6.0"),
        # ("github.com/golangci/golangci-lint/cmd/golangci-lint", "v1.54.2"),
        ("github.com/dmarkham/enumer", "v1.6.1"),
        ("mvdan.cc/gofumpt", "v0.5.0"),
        #("honnef.co/go/tools/cmd/staticcheck", "v0.4.6"),
        ("github.com/goreleaser/goreleaser", "v1.26.2"),
    ]
    for pkg, version in tools:
        run(f"go install {pkg}@{version}")


def task_mockgen() -> None:
    """生成 mock 文件"""
    dest = os.getenv("MOCK_DEST_FILE", "")
    pkg = os.getenv("PROJECT_PKG", "")
    iface = os.getenv("MOCK_INTERFACE", "")
    if not all([dest, pkg, iface]):
        print("Error: MOCK_DEST_FILE, PROJECT_PKG, and MOCK_INTERFACE are required")
        sys.exit(1)
    run(f"mockgen -destination test/{dest} -package test --build_flags=--mod=mod {pkg} {iface}")


def task_test_unit() -> int:
    """运行单元测试（忽略 legacy 包）"""
    return run("go test -short -timeout 29s ./...")


def task_test_race() -> None:
    """运行竞态检测"""
    run("go test -race ./...")


def task_test_e2e() -> None:
    """运行端到端测试"""
    # 检查 e2e 目录下是否有 Go 测试文件
    result = subprocess.run(
        "go list -tags=e2e ./e2e/...",
        shell=True, capture_output=True, text=True
    )
    packages = result.stdout.strip()
    if not packages:
        print("跳过 e2e 测试：没有找到测试包")
        return
    run(f"go test -tags=e2e {packages} -v")


def task_test_compat() -> None:
    """运行兼容性测试"""
    # 检查 compat 目录下是否有 Go 测试文件
    result = subprocess.run(
        "go list ./e2e/compat/...",
        shell=True, capture_output=True, text=True
    )
    packages = result.stdout.strip()
    if not packages:
        print("跳过 compat 测试：没有找到测试包")
        return
    run(f"go test {packages} -v -timeout 30m")


def task_lint() -> None:
    """运行代码检查"""
    # 检查 golangci-lint 是否安装
    result = run_quiet("which golangci-lint")
    if result != 0:
        print("请先安装 golangci-lint: go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest")
        sys.exit(1)
    run("golangci-lint run")


def task_clean() -> None:
    """清理构建产物"""
    run("rm -rf build")
    run("rm -f coverage.out coverage.html")
    run("go clean")


def task_generate() -> None:
    """生成代码（执行 go:generate 指令）"""
    run("go generate ./...")


def task_check_coverage() -> None:
    """检查测试覆盖率是否达到 100%"""
    Path("build").mkdir(exist_ok=True)
    # 使用 -tags=test 排除 main 函数（因为 main 函数无法直接测试）
    run("go test ./... -tags=test -coverprofile=build/coverage.out -covermode=atomic")
    print("检查覆盖率...")
    # 使用 go tool cover 检查总覆盖率
    result = subprocess.run(
        "go tool cover -func=build/coverage.out | grep total",
        shell=True, capture_output=True, text=True
    )
    if result.returncode != 0:
        print("无法获取覆盖率信息")
        sys.exit(1)

    coverage_line = result.stdout.strip()
    print(coverage_line)

    # 检查是否为 100%
    if "100.0%" not in coverage_line:
        print(f"覆盖率不足 100%")
        sys.exit(1)
    else:
        print("覆盖率达标: 100%")


def task_test_all() -> None:
    """运行所有测试"""
    tasks = [
        ("单元测试", task_test_unit),
        #("集成测试", task_test_integration),
        #("E2E 测试", task_test_e2e),
    ]

    for i, (name, fn) in enumerate(tasks, 1):
        print(f"\n=== {i}/{len(tasks)} 运行{name} ===")
        ret = fn()
        if ret != 0:
            print(f"{name}失败")
            sys.exit(1)

    print("\n=== 所有测试通过 ===")


# 任务映射表
TaskName = Literal[
    "mod", "check", "test", "build", "build-all", "release",
    "init", "mockgen", "test-unit", "test-race", "test-e2e",
    "test-compat", "test-all", "lint", "clean", "generate",
    "check-coverage"
]

TASKS: dict[TaskName, Callable[[], int | None]] = {
    "mod": task_mod,
    "check": task_check,
    "test": task_test,
    "build": task_build,
    "build-all": task_build_all,
    "release": task_release,
    "init": task_init,
    "mockgen": task_mockgen,
    # "env-up": task_env_up,
    # "env-down": task_env_down,
    # "env-status": task_env_status,
    "test-unit": task_test_unit,
    "test-race": task_test_race,
    "test-e2e": task_test_e2e,
    "test-compat": task_test_compat,
    # "test-integration": task_test_integration,
    "test-all": task_test_all,
    "lint": task_lint,
    "clean": task_clean,
    "generate": task_generate,
    "check-coverage": task_check_coverage,
}


def main() -> None:
    task: TaskName | str = sys.argv[1] if len(sys.argv) > 1 else "build"

    if task not in TASKS:
        print(f"Unknown task: {task}")
        print(f"Available tasks: {', '.join(TASKS.keys())}")
        sys.exit(1)

    sys.exit(TASKS[task]() or 0)


if __name__ == "__main__":
    main()
