# Focogram
Focogram是一个基于golang开发的社交聊天网站

环境要求
- golang 1.24
- mysql 8.0
- redis 6.0

启动教程
1.数据库配置
   - 创建数据库focogram
   - 在config/config.yml中配置数据库连接信息
    dsn: "<db_user>:<db_password>@tcp(127.0.0.1:3306)/focogram?charset=utf8mb4&parseTime=True&loc=Local"
2.项目启动
   - 进入项目目录
   - 执行命令 go run .
   - 进入在focogram/web目录
   - 执行命令 npm run dev
   - 项目启动后，浏览器会自动打开http://localhost:5173
3.docker启动
   - 进入项目目录
   - 执行命令 docker-compose up -d
   - 项目启动后，浏览器会自动打开http://localhost:5173