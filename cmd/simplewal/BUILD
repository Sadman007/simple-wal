load("@rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "wal",
    srcs = ["main.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//internal/wal:wal_lib",
        "//proto:wal_go_proto",
        "@org_golang_google_protobuf//proto",
    ],
)
