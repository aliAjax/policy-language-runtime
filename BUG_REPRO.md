# Bug Reproduction

向多个订阅者推送更新时，流可能提前结束、错误分支不关闭通道或因同步计数错误挂起。让多个订阅者并行接收，并注入校验失败即可复现。

现象：出现 WaitGroup panic、data race、少发更新或通道一直不关闭。
