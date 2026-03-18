# HTTP/2 Transport Mod

恢复自被删除的 HTTP 传输模块

只恢复了 HTTP/2 不过相对于原来支持了 H2C，服务端方面用几行代码顺带支持了 HTTP/1.1，但是客户端就懒得写了（个人觉得没意义）

你在 90% 的时候都可以去用 XHTTP 效果差不多。这里是我觉得它功能复杂+被一堆 AI 轮番轰炸后代码太乱才从以前的传输恢复并经由本人重写优化。

## HttpObject

`HttpObject` 对应传输配置的 `httpSettings` 项。

```json
{
  "network": "http",
  "httpSettings": {
    "host": ["example.com"],
    "path": "",
    "method": "POST",
    "idle_timeout": 45,
    "health_check_timeout": 15,
    "headers": {
      "My-Header": ["my-value"]
    }
  }
}
```

Host 被我改成单个了，配置格式保留数组形式但是只取第一个，客户端默认使用 dest 服务端设置后会检查发来的 Host。

Path 在在客户端会在请求里加上这个 Path。服务端会检查请求的 Path 是否以这个 Path 开头。

超过 idle_timeout 无任何数据发送会发出 ping 帧，如果 health_check_timeout 秒没回复则断开连接这。这其实不是 idle timeout 的意思，更类似 keepalive period，不过为了兼容旧配置而保留。
