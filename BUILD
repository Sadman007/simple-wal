load("@rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "wal_lib",
    srcs = [
        "internal/wal/wal.go",
    ],
    importpath = "simplewal/internal/wal",
    visibility = ["//visibility:public"],
)

go_binary(
    name = "simplewal",
    srcs = ["cmd/simplewal/main.go"],
    importpath = "simplewal/cmd/simplewal",
    visibility = ["//visibility:public"],
    deps = [":wal_lib"],
)

go_test(
    name = "wal_test",
    srcs = ["internal/wal/wal_test.go"],
    importpath = "simplewal/cmd/simplewal",
    visibility = ["//visibility:public"],
    deps = [":wal_lib"],
)
