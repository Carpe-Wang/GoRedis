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
  * 
