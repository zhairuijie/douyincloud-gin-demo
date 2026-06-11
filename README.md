本demo使用开源hertz生成，仅做参考, 创作者可根据文档实现相关接口即可

### serv_template_open目录说明
- biz // business 层，存放业务逻辑相关流程
    - component     // avatar原子，未来会提供通用原子实现，本期是简单demo
    - implements    // 用户实现入口
        - chat_stream.go    // 流式会话接口ChatStream实现入口
        - on_boarding.go    // 开场白流式接口OnBoarding实现入口
        - chat.go           // 非流式对话接口，通常在离线场景使用
    - convertor             // 结构体转换
    - dal                   // 远程访问
    - model                 // 结构体定义
    - router                // hertz 自动生成文件，接口路由/中间件
    - util                  // 工具包
        - resp_writer.go    // 简单封装http chunk返回的工具包
    - init.go               // 业务全局初始化
- script
    - bootstrap.sh // 本期启动脚本
- build.sh         // 本地编译脚本
- go.mod           // 包管理，创作者可以自定义module
- main.go               // 服务启动入口
- Dockerfile            // 抖音云部署需要；
- run.sh                // 抖音云部署需要