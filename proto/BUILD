load("@rules_go//proto:def.bzl", "go_proto_library")
load("@rules_proto//proto:defs.bzl", "proto_library")

go_proto_library(
    name = "wal_go_proto",
    importpath = "github.com/Sadman007/simplewal/proto",
    protos = [":wal_proto"],
    visibility = ["//visibility:public"],
)

proto_library(
    name = "wal_proto",
    srcs = ["wal.proto"],
    visibility = ["//visibility:private"],
)
