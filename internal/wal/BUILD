load("@rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "wal_lib",
    srcs = [
        "config.go",
        "entry.go",
        "segments.go",
        "wal.go",
    ],
    importpath = "github.com/Sadman007/simplewal/internal/wal",
    visibility = ["//:__subpackages__"],
    deps = [
        "//proto:wal_go_proto",
        "@org_golang_google_protobuf//proto",
    ],
)

go_test(
    name = "wal_test",
    srcs = ["wal_test.go"],
    deps = [":wal_lib"],
)
