# learning-dapr



## Questions

1. how to upgrade dapr without downtime on k8s (https://www.youtube.com/watch?v=HuJepZXng_c)



## Benefit

1. No SDK needed

1. run on k8s or on-premises

1. use Http to call GRPC

1. grpc load-balancing 

   https://github.com/dapr/dapr/issues/1444





### Bindings



### Cron

1. 需要寫 http verb options 來支援
1. 假設一個排成設定 5 秒跑一次, Dapr 會固定每 5 秒觸發, 不管上次的是否有成功的執行完成
1. 如果跑在 k8s 上面，如果這個服務有 2 個 pods, 這兩個 pods 同時都會收到觸發


