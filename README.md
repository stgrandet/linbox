# linbox

## Development:
This project use gpm to manage dependency versions.
In order to build this project, run command below:

```
gpw install
go build linbox/...

```
    

Build a test server and client under inmemory package.
Run commands below:

```

go build linbox/inmemory  // build 程序

./inmemory server   // 启动 server 端

./inmemory client 1 // 启动 client 端，并且指定用户 id  为 1

// 在 client 端连接上server 之后， 可以尝试输入如下命令:
SYNC_UNREAD_REQUEST  // 发送更新请求
READ_ACK_REQUEST 1_2 1  // 发送更新确认
PULL_OLD_MSG_REQUEST 1 100 10  // 拉取 user 1 从 100 开始向下的 10 个记录
SEND_MSG_REQUEST 2 hello // 向 user 2 发送信息 hello
```


