package main

import "fmt"
import "net"
import "bufio"
import "math/rand"
import "unicode/utf8"
import "unicode"
import "strings"
import "time"
import "os"
import "strconv"
import "flag"

func main() {
    argsIpPort := flag.String("ip","","a string") //"127.0.0.1:9999"
    argsPort := flag.String("port",":9999","a string") //":9999"
    argsMaxClients := flag.Int("n",2,"an int") //"10"
    flag.Parse()

    if(*argsIpPort == "") {
      fmt.Println("Server launched...")
      // listen on all interfaces
      ln, _ := net.Listen("tcp", ":"+*argsPort)
      i := 0
      for {
          // accept connection on port
          if i < *argsMaxClients {
            c, _ := ln.Accept()
            i++
            fmt.Printf("Client %v connected...", i)
            fmt.Println("---")
            // handle the connection
            go handleServerConnection(c, i, &i)
          }
      }
    } else {
      i := 0 // connection index
      hash := ""
      key := ""
      var s Session_protector
      // connect to server
      for {
      connect:
          c, errConn := net.Dial("tcp", *argsIpPort)
          if errConn != nil {
              continue
          } else {
              i++
              if i <= 1 {
                keyServer := ""
                keysMatched := 0  //how much keys are the same

                scannerServ := bufio.NewScanner(c)
                fmt.Println("Connected to server...")
                fmt.Println("---")
                //giving hash to server
                hash = get_hash_str()
                s.__hash = hash
                fmt.Println(hash)
                fmt.Fprintf(c, hash+"\n")
                scannerServ.Scan()
                fmt.Println("Server sends: " + scannerServ.Text())
                //giving key to server
                key = get_session_key()
                fmt.Println(key)
                fmt.Fprintf(c, key+"\n")
                scannerServ.Scan()
                fmt.Println("Server sends: " + scannerServ.Text())
                //comparing next calculated keys
                for j:=1; j<7; j++ {
                    key = s.next_session_key(key)
                    fmt.Fprintf(c, key+"\n")

                    fmt.Println("Client key: " + key)
                    fmt.Println("Server key: " + keyServer)
                    key = strings.TrimRight(key, "\n")
                    keyServer = strings.TrimRight(keyServer, "\n")
                    if key == keyServer {
                      fmt.Println("Keys matched")
                      keysMatched = keysMatched + 1
                    }

                    if keysMatched == 5 {
                      fmt.Println("Client and server keys are the same. It's safe!")
                      break
                    }
                    scannerServ.Scan()
                    keyServer = scannerServ.Text()
                    fmt.Println("-----------------------------")
                }
              }
          }

          for {
              // read in input from stdin
              scannerStdin := bufio.NewScanner(os.Stdin)
              fmt.Print("Send message: ")
              for scannerStdin.Scan() {
                  text := scannerStdin.Text()
                  fmt.Println("---")
                  // send to server
                  _, errWrite := fmt.Fprintf(c, text+"\n")
                  if errWrite != nil {
                      fmt.Println("Server offline, attempting to reconnect...")
                      goto connect
                  }
                  fmt.Print("Server receives: " + text)
                  if text == "exit" {
                    os.Exit(0)
                  }
                  break
              }
              fmt.Println("---")
          }
      }
    }
}
func handleServerConnection(c net.Conn, i int, clientCount *int) {
    iteration := 0
    nextKey := ""
    var s Session_protector
        // scan message
    scanner := bufio.NewScanner(c)

    for scanner.Scan() {
        msg := scanner.Text()
        fmt.Printf("Client %v sends: %v", i, msg)
        fmt.Println("\n")
        if iteration == 0 {
          s.__hash = msg
          c.Write([]byte(msg + "\n"))
          iteration = iteration + 1
        } else if iteration < 7 {
          nextKey = s.next_session_key(msg)
          c.Write([]byte(nextKey + "\n"))
          fmt.Printf("Client %v receives: %v", i, nextKey)
          fmt.Println("\n")
          iteration = iteration + 1
        }
        if scanner.Text() == "exit" {
          *clientCount = *clientCount - 1
          fmt.Println("client disconnected")
          c.Close()
        }
    }
}

type Session_protector struct {
	__hash string
}

func get_session_key() string {
  rand.Seed(time.Now().UTC().UnixNano())
	var result = ""
	for i := 1; i < 11; i++ {
		result += strconv.Itoa(rand.Intn(9))
	}
	return result
}

func get_hash_str() string {
  rand.Seed(time.Now().UTC().UnixNano())
	var li = ""
	for i := 0; i < 6; i++ {
		li += strconv.Itoa(rand.Intn(6))
	}
	return li
}


func (s Session_protector) __calc_hash(session_key string, val int) int {
	var result = 0
	for i := 0; i < utf8.RuneCountInString(session_key); i++ {
		current, err := strconv.Atoi(string(session_key[i]))
		if err != nil {

		} else {
			if val == 1 {
				result = result + current ^ 16
			} else if val == 2 {
				result = result + current * 50
			} else if val == 3 {
				result = result + current ^ 32
			} else if val == 4 {
				result = result + current * 30
			} else {
        result = result + current * 123
      }
		}
	}
	return result
}

func (s Session_protector) next_session_key(session_key string) string {
	var result = 0
	if s.__hash == "" {
		fmt.Println("It's empty")
	} else {
		for _, element := range s.__hash {
			if !unicode.IsDigit(element) {
				fmt.Println("Symbol is not digit!")
			}
		}
		for i := 0; i < utf8.RuneCountInString(s.__hash); i++ {
			result += s.__calc_hash(session_key, i)
		}
	}
	return (strconv.Itoa(result))
}
