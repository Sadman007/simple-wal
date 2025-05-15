load("@rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_binary(
    name = "simplewal",
    srcs = ["cmd/simplewal/main.go"],
    importpath = "simplewal/cmd/simplewal",
    visibility = ["//visibility:public"],
    deps = ["//internal/wal:wal_lib"],
)