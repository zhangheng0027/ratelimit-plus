# ratelimit
```go
import "github.com/zhangheng0027/ratelimitplus"
```
对 [ratelimit](github.com/juju/ratelimit) 项目进行增强，增加了上游限流控制

上游控制方式有两种

### 方式1：串行控制，默认方式

当前最大速度是小于所有上游速度的最小值，类似与水管串联，最小的那个水管决定了最终水流速度

```go
b1 := NewBucketWithRate(10, 100)
b2 := NewBucketWithRate(8, 100)
b3 := NewBucketWithRate(7, 100)

b1.AddUpstream(b2)
b1.AddUpstream(b3)

// b1 的上游是 b2 和 b3, b1 最终速度是 7
```

```go
b1 := NewBucketWithRate(10, 100)
b2 := NewBucketWithRate(8, 100)
b3 := NewBucketWithRate(7, 100)

b1.AddUpstream(b2)
b2.AddUpstream(b3)

// b1 的上游是 b2, b2 的上游是 b3, b2 的最终速度为 7, b1 的最终速度为 7
```

### 方式2：并行控制
当前最大速度是所有上游速度的和，类似与水管并联，所有水管的速度之和决定了最终水流速度

