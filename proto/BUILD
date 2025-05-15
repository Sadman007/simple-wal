load("@rules_proto//proto:defs.bzl", "proto_library")
load("@rules_go//proto:def.bzl", "go_proto_library")

go_proto_library(
    name = "wal_go_proto",
    protos = [":wal_proto"],
    importpath = "simplewal/proto",
    visibility = ["//visibility:public"],
)

proto_library(
    name = "wal_proto",
    srcs = ["wal.proto"],
    visibility = ["//visibility:public"],
)