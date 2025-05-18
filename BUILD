load("@rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "wal",
    srcs = ["cmd/simplewal/main.go"],
    visibility = ["//visibility:public"],
    deps = ["//internal/wal:wal_lib"],
)