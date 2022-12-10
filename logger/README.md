<div align="center">
  <p>
      <pre style="float:center">
                                                     ('-.    _  .-')   
                                                   _(  OO)  ( \( -O )  
 ,--.       .-'),-----.   ,----.       ,----.     (,------.  ,------.  
 |  |.-')  ( OO'  .-.  ' '  .-./-')   '  .-./-')   |  .---'  |   /`. ' 
 |  | OO ) /   |  | |  | |  |_( O- )  |  |_( O- )  |  |      |  /  | | 
 |  |`-' | \_) |  |\|  | |  | .--, \  |  | .--, \ (|  '--.   |  |_.' | 
(|  '---.'   \ |  | |  |(|  | '. (_/ (|  | '. (_/  |  .--'   |  .  '.' 
 |      |     `'  '-'  ' |  '--'  |   |  '--'  |   |  `---.  |  |\  \  
 `------'       `-----'   `------'     `------'    `------'  `--' '--' 
  </pre>
  </p>
  <p>
<br> 
</p>


[![Build Status](https://github.com/wwqdrh/gokit/logger/actions/workflows/push.yml/badge.svg)](https://github.com/wwqdrh/gokit/logger/actions)
[![codecov](https://codecov.io/gh/wwqdrh/logger/branch/main/graph/badge.svg?token=G4DUWSQHPJ)](https://codecov.io/gh/wwqdrh/logger)

  </p>
</div>

# Usage

基于zap封装的日志库

1、可以配置同时写入的日志文件
2、日志文件进行了切割
3、输出在控制台中的日志做了高亮处理

- 1、动态调整日志级别(这样就可以在线上也可以跳出打印debug信息)

默认监听的环境变量名为`LOG_LEVEL`

## trace id

```go
// middleware
func AddTraceId() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traceId := ctx.GetHeader("traceId")
		if traceId == "" {
			traceId = uuid.New().String()
		}
		ctx, log := logger.DefaultLogger.AddCtx(ctx.Request.Context(), zap.Any("traceId", traceId))
		ctx.Request = ctx.Request.WithContext(ctx)
		log.Info("AddTraceId success")
		ctx.Next()
	}
}

// handler
func handler(ctx *gin.Context) {
  l := logger.DefaultLogger.GetCtx(ctx.Request.Context())
  l.Info("a info") // with a trace id
  l.Debug("a debug") // with a trace id
}
```