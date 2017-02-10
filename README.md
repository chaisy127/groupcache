# groupcache中文注释

## 摘要

groupcache是一个缓存及缓存过滤库，作为 memcached 许多场景下的替代版本

API文档和例子详见http://godoc.org/github.com/golang/groupcache

## 对比memcached

### **与memcached相同的地方**:

 * 通过key分片，并且通过key来查询响应的peer

### **与memcached不同的地方**:

 * 不需要对服务器进行单独的设置，这将大幅度减少部署和配置的工作量。
   groupcache既是客户端库也是服务器库，并连接到自己的peer上。

 * 具有缓存过滤机制。
   众所周知，在memcached出现“Sorry，cache miss（缓存丢失）”时，
   经常会因为不受控制用户数量的请求而导致数据库（或者其它组件）产生“惊群效应（thundering herd）”；
   groupcache会协调缓存填充，只会将重复调用中的一个放于缓存，而处理结果将发送给所有相同的调用者。

 * 不支持多个版本的值。
   如果“foo”键对应的值是“bar”，那么键“foo”的值永远都是“bar”。
   这里既没有缓存的有效期，也没有明确的缓存回收机制，因此同样也没有CAS或者Increment/Decrement。

 * ... 基于上一点的改变，groupcache 就具备了自动备份“超热”项进行多重处理，
   这就避免了 memcached 中对某些键值过量访问而造成所在机器 CPU 或者 NIC 过载。

 * 当前只支持Go语言。我(bradfitz@)不太可能实现所有语言的接口代码

## 运行机制

简而言之，groupcache查找一个Get（“foo”）的过程类似下面的情景

（机器#5上，它是运行相同代码N台机器集中的一台）：

 1. key“foo”的值是否会因为“过热”而储存在本地内存，如果是，就直接使用。

 2. key“foo”的值是否会因为peer #5是其拥有者而储存在本地内存，如果是，就直接使用。

 3. 首先确定key “fool”是否归属自己N个机器集合的peer中，
    如果是，就直接加载。如果有其它的调用者介入（通过相同的进程或者是peer的RPC请求，
    这些请求将会被阻塞，而处理结束后，他们将直接获得相同的结果）。
    如果不是，将key的所有者RPC到相应的peer。如果RPC失败，那么直接在本地加载（仍然通过备份来应对负载）。

## 用户

groupcache已经在dl.Google.com、Blogger、Google Code、Google Fiber、Google生产监视系统等项目中投入使用。

## 演示

见http://talks.golang.org/2013/oscon-dl.slide

## 帮助

在golang-nuts邮件列表中讨论或者提问。
