package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {

	g, ctx := errgroup.WithContext(context.Background())
	servers := make([]*http.Server, 0, 3)

	//拦截linux signal信号量
	g.Go(func() error {

		defer func() {
			//防止野生goroutine的panic
			if r := recover(); r != nil {
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				fmt.Printf("errgroup: panic recovered: %s\n%s", r, buf)
			}
			//关停服务
			shutDown(servers)
		}()

		ch := make(chan os.Signal)
		for {
			//ctrl + c, ctrl + \, 程序结束 , ctrl + z
			signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGTSTP)

			select {
			case <-ctx.Done():
				fmt.Println("正常结束服务")
				return nil
			case sig := <-ch:
				fmt.Println("收到退出信号：", sig)
				return nil
			}
		}
	})

	//开启服务
	for i := 0; i < 3; i++ {
		server := &http.Server{Addr: fmt.Sprintf(":808%d", i), Handler: MyHandler{}}
		servers = append(servers, server)

		g.Go(func() error {
			defer func() {
				//防止野生goroutine的panic
				if r := recover(); r != nil {
					buf := make([]byte, 64<<10)
					buf = buf[:runtime.Stack(buf, false)]
					fmt.Printf("errgroup: panic recovered: %s\n%s", r, buf)
				}
			}()

			if err := server.ListenAndServe(); err != nil {
				fmt.Printf("错误退出 (%+v)", err)
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		//有一个服务挂掉就全部退出
		shutDown(servers)
	}

}

type MyHandler struct {
}

func (handler MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "你好呀")
}

func shutDown(servers []*http.Server) {
	shoutDownCtx, _ := context.WithTimeout(context.Background(), time.Second)
	for _, server := range servers {
		server.Shutdown(shoutDownCtx)
	}
}
