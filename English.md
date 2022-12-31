# GoRedis(This project is designed for my graduation from Henan University of Technology in 2023)
* The program is temporarily open because it is used to apply for American graduate courses.
### Project introduction
* Use Go to implement basic Redis commands, such as set, get, AOF and other basic functions.
### Method of Application
* Standalone Mode
  * Start the program directly.
  * Or use the `go build ` command to generate an executable file
* Cluster startup
  * Use the `go build` command to generate an executable file.
  * Put the executable file in a separate folder. Note: The folder needs to contain the configuration (redis. conf).
  * The port cannot be duplicate. Self is its own IP+port number, and peers is the other IP+port number in the cluster.
  * like this：
    ```
    bind 0.0.0.0
    port 6379

    appendonly yes
    appendfilename appendonly.aof

    self 127.0.0.1:6379
    peers 127.0.0.1:6380
    ```
  * Then macOS or Linux enters the corresponding folder to execute `/ GoRedis` can be run.
  * After running, you can see`[INFO][server.go:42] 2022/10/09 10:36:47 bind: 0.0.0.0:6380, start listening...`。
  * Connect TCP client through network debugging assistant`[INFO][server.go:71] 2022/10/09 10:37:30 accept link`。
  * The RESP command can be sent and accept normally 
### log
* The log information mainly records the project start time.
* Whether there is connection, connection time, etc.
### AOF
* The operation information is mainly recorded through appendonly.aof under the project.
* When restarting the system, conduct LoadAof to prevent the loss of restart information.

### PS：It is easy to look at the Resp protocol first to facilitate sending requests later
>*3/r/n$3/r/nset/r/n$4/r/nname/r/n$5/r/npdudo/r/n

> Because I use macOS, I can't directly use the network debugging assistant to send tcp when using RESP. I can only use the telnet command.
>
>The key point is that you need to download Homebrew first before you can execute commands on the mac terminal.

