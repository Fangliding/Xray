# Xray

自用 主要用于解除一些个人认为不合适的限制或者加一些不合适主线的功能

同步主线的时间不定 根据心情 有时候有看起来比较灵车的 commit 可能会暂时不同步 不过全部
上游的 main 分支会由 bot 每天拉到 [upstream](https://github.com/Fangliding/Xray/tree/upstream) 分支 可以在那里观察和上游的差异

这里的 patch 尽可能保持简单好处理 就像我之前坚持的那样 便于 rebase 同步上游 有需要的可以自行使用这里的补丁进行 rebase 或者 cherry-pick 应该不太容易出现冲突 但是为了这个目的要保持 git 整洁所以 force push 是经常的事情

## added

如果证书只填了证书文件路径 没填私钥文件路径文件 自动在同文件夹下扫描匹配的私钥文件 非常不安全 非常好用（我实在懒得把路径复制一遍然后把 fullchain.cer 改成 xxx.key）

## modded

加回 allowInsecure

XHTTP 无需 padding 而且默认也不会加 padding

移除使用 VMESS 等功能时乱七八糟的 deprecated 警告

允许 vsion 和 XHTTP 使用 mux cool （XHTTP 有点小用 但是 vision + mux 不如不开 vision）

## other

上游目前有的的地方已经成为 vibe 游乐场 重灾区有 API（大量为了机场用途而添加的鬼玩意）XHTTP Finalmask 反正无人在意 放弃治疗了 现在 AI 确实强大 这些功能不代表用不了 但是大量没什么大用的功能让本来的屎山越来越糟糕也是没办法的
