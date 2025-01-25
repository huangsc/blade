package middleware

import (
	"context"
)

// Handler 定义中间件处理函数
type Handler func(ctx context.Context, req interface{}) (interface{}, error)

// Middleware 定义中间件函数类型
type Middleware func(Handler) Handler

// Chain 将多个中间件串联成一个
func Chain(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Recovery 定义恢复中间件
func Recovery() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req interface{}) (resp interface{}, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = recoverError(r)
				}
			}()
			return next(ctx, req)
		}
	}
}

// recoverError 将 panic 转换为 error
func recoverError(r interface{}) error {
	if err, ok := r.(error); ok {
		return err
	}
	return nil
}
