# learning-dapr

1. No SDK
1. use Http to call GRPC

dapr run --app-id order --app-port 3000 --port 3500 go run .


## Bindings

### Cron

1. 需要寫 http verb options 來支援
1. 假設一個排成設定 5 秒跑一次, Dapr 會固定每 5 秒觸發, 不管上次的是否有成功的執行完成
1. 如果跑在 k8s 上面，如果這個服務有 2 個 pods, 這兩個 pods 同時都會收到觸發


