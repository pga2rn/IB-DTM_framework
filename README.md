# Simulator implementation

## project Architecture:

1. config
2. core
3. dtm
4. rpc
5. sim-map

6. statistics

7. vehicle


## package definition

### core

1. 负责simulation的逻辑
2. simulation初始化:
   1. 初始化session
   2. 初始化vehicles
   3. 初始化RSU
3. simulation routine
   1. 等待genesis
   2. genesis达成
4. trust value计算: 直接通过生成的trustvalueoffset来计算trust value

### dtm

#### rsu

1. rsu初始化逻辑

   1. 数据结构初始化

   2. 通过grpc连接到外部的RSU module

2. rsu逻辑(在simulator内的)

   主要是给simulator执行rsu逻辑的

   1. processslot, 更新simulator内的rsu的数据结构, 每个RSU得到的trust value
   
3. 提供一个rpc接口供外部rsu module来获得当前可用的trust value offset值

### rpc

1. 主要是提供数据查询api,

   提供如下信息:

   1. 当前session情报
      1. epoch, slot
      2. vehicles nums
      3. compromised rsu portion, pos





# 最简易实验搭建

1. 各个数据结构都准备好
   1. sim session
   2. config
   3. map
   4. rsu
   5. vehicles

2. 简易实验流程
   1. dtm/rsu不连接到外部rsu模块
   2. genesis自定, 不和外部eth2模块进行通信
   3. 不引入compromisedRSU
   4. 直接把sim跑起来



# 实验metrics

1. efficiency
   1. time efficiency by % (已经upload的和没有upload的trust value offset的比值)
2. bias
   1. bias caused by delay (和simulation里生成的accurate trust value进行比对)
   2. 和sim里生成的即时bias trust value进行对比