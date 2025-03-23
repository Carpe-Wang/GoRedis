### Project Overview
* This project implements basic Redis commands using Go, such as `SET`, `GET`, and AOF (Append Only File) persistence.

### How to Use

* **Single-node Mode**
  * Simply run the program directly.

* **Cluster Mode**
  * Build the executable using the `go build` command.
  * Place the executable into a separate folder. **Note:** The folder must include the configuration file (`redis.conf`).
  * Ports must be unique. `self` refers to the current node's IP and port, while `peers` refers to other nodes in the cluster.
  * Example `redis.conf`:
    ```
    bind 0.0.0.0
    port 6379

    appendonly yes
    appendfilename appendonly.aof

    self 127.0.0.1:6379
    peers 127.0.0.1:6380
    ```
  * On macOS or Linux, navigate to the corresponding folder and run `./goRedis`.

  * After running, you'll see logs like:
    ```
    [INFO][server.go:42] 2022/10/09 10:36:47 bind: 0.0.0.0:6380, start listening...
    ```
  * Connect via TCP using a network debugging tool. Successful connection is logged:
    ```
    [INFO][server.go:71] 2022/10/09 10:37:30 accept link
    ```
  * You can then send commands as expected.

### Logging
* Logs mainly record project start time.
* Also tracks connection events, such as whether a connection was established and when.

### AOF (Append Only File)

* Operations are logged in `appendonly.aof` located in the project directory.
* During system restart, `LoadAof` ensures data is restored, preventing data loss.

### Note: Learn RESP Protocol First
To better send requests later, get familiar with the RESP protocol:
* \r\n$3\r\nset\r\n$4\r\nname\r\n$5\r\npdudo\r\n

### Note for macOS
> Since I use macOS, I couldn’t use a network debugging tool to directly send TCP data with RESP.  
> Instead, I used the `telnet` command in the terminal.  
> Make sure to install Homebrew first to enable the command.

---

## Project Demonstration

* **Startup**
  * ![img.png](img/img.png)
* **SET Command Test**
  * ![img_1.png](img/img_1.png)
* **AOF Output**
  * ![img_3.png](img/img_3.png)
* **GET Command Test**
  * ![img_2.png](img/img_2.png)

---

## Graduation Project Preparation (For reference only)

### Why I Chose This Topic
* Inspired by how HTTP/3 no longer uses TCP as the transport layer, I wanted to record and understand TCP in my own way.
* Redis plays a crucial role in backend software development, second only to traditional databases. It also helps reduce I/O latency caused by frequent database access.

### Why Redis is Fast
* According to Computer Architecture principles, memory access speed hierarchy is: Register < L1/L2/L3 Cache < RAM < SSD < HDD.
  Redis stores data in memory and persists it on disk, making it faster than traditional databases.
* Redis uses I/O multiplexing, meaning a single thread can efficiently handle many connections. Most operations are memory-based, and memory is not a performance bottleneck.
* Being in-memory and working like a HashMap, Redis achieves O(1) time complexity for reads and writes.

### Difference from Traditional Databases
* Redis is a non-relational key-value store used for massive data handling.
* NoSQL databases store data in memory (cache), whereas relational databases store data on disk, resulting in slower query speeds.
* NoSQL is key-value based and avoids SQL parsing, leading to very high performance.

---

## Design Details

### Lightweight TCP Server

* TCP configurations can be found in `config.config` and `redis.conf`.

```go
func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		_ = conn.Close()
	}

	client := &EchoClient{
		Conn: conn,
	}
	h.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("connection close")
				h.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}
```
