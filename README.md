# go-crontab
golang分布式调度系统

#项目结构：
###go-crontab
    /master
    /worker
    /common   (工具包)

##master:
1.搭建go项目框架，配置文件，命令行参数，线程配置...
2.给web后台提供http API，用于管理job
3.写一个web后台的前端页面，boostrap+jquery，前后端分离
