package main

import (
	"bufio"
	"fmt"
	"github.com/hashicorp/consul/api"
	"math/rand"
	"net"
	"net/url"
	"os"
	serviceDiscovery "serviceDiscovery/consul"
	"strconv"
	"strings"
	"sync"
	"time"
)

func consulTest() {
	cl, _ := api.NewClient(api.DefaultConfig())
	he := cl.Health()
	client, _ := serviceDiscovery.NewConsulClient(he)
	qo := api.QueryOptions{
		Namespace:         "",
		Datacenter:        "",
		AllowStale:        false,
		RequireConsistent: true,
		UseCache:          false,
		MaxAge:            0,
		StaleIfError:      0,
		WaitIndex:         0,
		WaitHash:          "",
		WaitTime:          0,
		Token:             "",
		Near:              "",
		NodeMeta:          nil,
		RelayFactor:       0,
		LocalOnly:         false,
		Connect:           true,
		Filter:            "",
	}
	//services, _, _ := client.Services(&qo)
	//fmt.Println(services)
	////keys := make([]string, len(services))
	//
	//i := 0
	//for k := range services {
	//	keys[i] = k
	//	i++
	//}
	//fmt.Println(keys)

	dcs := []string{"dc1", "local-dc"}
	tags := []string{"golang"}
	serviceName := "web"
	namespace := ""
	//_, _, _ = client.Nodes(dcs,"web","",tags,&qo)
	//for k, _ := range services {
	//	fmt.Println(k)
	//	s := reflect.ValueOf(qm).Elem()
	//	typeOfT := s.Type()
	//
	//	for i := 0; i < s.NumField(); i++ {
	//		f := s.Field(i)
	//		fmt.Printf("%d: %s %s = %v\n", i,
	//			typeOfT.Field(i).Name, f.Type(), f.Interface())
	//	}
	//	fmt.Println("Err",err)
	//
	//}
	//dcs := []string{"dc1", "local-sdc"}
	//tags := []string{"tag1", "golang"}
	//services, _, _ := client.Nodes(dcs, "", tags, &qo)
	//fmt.Println(services)

	//ticker := time.NewTicker(1 * time.Second)
	//quit := make(chan struct{})
	//go func() {
	//	for {
	//		select {
	//		case <-ticker.C:
	//			// do stuff
	//			//todo wrap in a context
	//			_, _, _ = client.Nodes(dcs, "web", "", tags, &qo)
	//		case <-quit:
	//			ticker.Stop()
	//			return
	//		}
	//	}
	//}()
	//time.Sleep(6 * time.Second)

	hc, _, _ := he.Service(serviceName, tags[0], true, &qo)
	for _, i2 := range hc {
		fmt.Println(i2.Service.Port)
	}

	consulWatcher, _ := serviceDiscovery.NewConsulWatcher(client)
	query := serviceDiscovery.Query{
		QString: serviceDiscovery.QueryString{
			Datacenters: dcs,
			ServiceName: serviceName,
			Namespace:   namespace,
			Tags:        tags,
		},
		QueryOptions: &qo,
	}
	doneChan := make(chan int)
	nodeInfoChan, errChan := consulWatcher.Watch(&query, doneChan)
	go func() {
		select {
		case <-time.After(120 * time.Second):
			doneChan <- 1
		}

	}()
	for {
		select {
		case n, ok := <-nodeInfoChan:
			if !ok {
				return
			}
			fmt.Println("nodeInfo chan:", n)
		case e, ok := <-errChan:
			if !ok {
				return
			}
			fmt.Println("errrchan:", e)
		}

	}

}
func channelTest() {
	c1 := make(chan string)
	c2 := make(chan string)
	reader := bufio.NewReader(os.Stdin)

	go func() {
		for true {
			time.Sleep(1 * time.Second)
			text, _ := reader.ReadString('\n')
			if strings.Contains(text, "h") {
				c1 <- text
			} else {
				c2 <- text
			}

		}
	}()

	for {
		select {
		case msg1 := <-c1:
			fmt.Println("received c1", msg1)
		case msg2 := <-c2:
			fmt.Println("received c2", msg2)
		}

	}
}
func boring(j int, output chan<- string) {
	for i := 0; ; i++ {
		output <- "msg" + strconv.Itoa(j) + " " + strconv.Itoa(i)
		//fmt.Println("msg", j)
		time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
	}
}
func boringRec(inp <-chan string) {
	for {
		select {
		case v := <-inp:
			fmt.Println(v)
		}
	}
}
func testBoring() {
	out := make(chan string)
	for i := 0; i < 2; i++ {
		go boring(i, out)
	}
	timeout := time.After(2 * time.Second)
	for {
		select {
		case val := <-out:
			fmt.Println(val)
		case <-timeout:
			fmt.Println("Time after ")
			return
		}
	}

}
func testChanClose() {
	c := make(chan string)
	go func() { boring(1, c) }()
	go func() { boringRec(c) }()
	time.Sleep(3 * time.Second)
	close(c)
	fmt.Println("closed")
}
func fetchAll() error {
	var N = 4
	quit := make(chan bool)
	errc := make(chan error)
	done := make(chan error)
	for i := 0; i < N; i++ {
		go func(i int) {
			// dummy fetch
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			err := error(nil)
			if rand.Intn(2) == 0 {
				err = fmt.Errorf("goroutine %d's error returned", i)
			}
			ch := done // we'll send to done if nil error and to errc otherwise
			if err != nil {
				ch = errc
			}
			select {
			case ch <- err:
				return
			case <-quit:
				return
			}
		}(i)
	}
	count := 0
	for {
		select {
		case err := <-errc:
			close(quit)
			return err
		case <-done:
			count++
			if count == N {
				return nil // got all N signals, so there was no error
			}
		}
	}
}
func pipeline() {
	genFun := func(val ...int) <-chan int {
		out := make(chan int)
		go func() {
			for _, v := range val {
				fmt.Println(v, "from gen")
				out <- v
			}
			close(out)
		}()
		return out
	}
	sqFun := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			for v := range in {
				out <- v * v
				fmt.Println(v*v, "from sq")
			}
			close(out)
		}()
		return out
	}
	merge := func(cs ...<-chan int) <-chan int {
		var wg sync.WaitGroup
		out := make(chan int)

		// Start an output goroutine for each input channel in cs.  output
		// copies values from c to out until c is closed, then calls wg.Done.
		output := func(c <-chan int) {
			for n := range c {
				out <- n
			}
			wg.Done()
		}
		wg.Add(len(cs))
		for _, c := range cs {
			go output(c)
		}

		// Start a goroutine to close out once all the output goroutines are
		// done.  This must start after the wg.Add call.
		go func() {
			wg.Wait()
			close(out)
		}()
		return out
	}
	in := genFun(2, 3, 4)
	c1 := sqFun(in)
	//c2 := sqFun(in)
	_ = merge(c1)
	//for i := range merge(c1,c2) {
	//	fmt.Println(i)
	//}
	for i := range c1 {
		fmt.Println(i)
	}

}
func chanExitAfter() {
	c := make(chan int)
	go func(p chan int) {
		for i := 0; i < 5; i++ {
			p <- i
			//time.Sleep(1*time.Second)
		}
		close(p)
	}(c)
	w := time.After(2 * time.Second)
	for {
		select {
		case v, ok := <-c:
			if !ok {
				c = nil
				break
			}
			fmt.Println(v)
		case <-w:
			//closing after 2 sec
			fmt.Println("b")
			return
		}
	}

}
func doneChan() {
	produce := func(ch, done chan int, data []int) {
		defer func() { done <- 1 }()
		for _, v := range data {
			ch <- v
		}
	}
	mins := make(chan int)
	maxs := make(chan int)
	done := make(chan int)
	go produce(mins, done, []int{2, 5, 1, 4})
	go produce(maxs, done, []int{4, 8, 9, 7})
	for n := 2; n > 0; {
		select {
		case p := <-mins:
			fmt.Println("Min:", p) //consume output
		case p := <-maxs:
			fmt.Println("Max:", p) //consume output
		case <-done:
			n--
		}
	}
}
func pingPong() {
	type Ball struct {
		hits int
	}
	player := func(name string, table chan *Ball) {
		for {
			ball := <-table
			ball.hits++
			fmt.Println(name, ball.hits)
			time.Sleep(100 * time.Millisecond)
			table <- ball
		}
	}
	table := make(chan *Ball)
	go player("ping", table)
	go player("pong", table)
	//table <- new(Ball)
	//time.Sleep(1 * time.Second)
	<-table //game over
}
func urlTest() (bool, error) {
	urls := []string{
		"",
		"http://www.dumpsters.com",
		"https://www.dumpsters.com:443",
		"testing-path.com",
		"abc.com:80",
		"http://abc.com:80",
		"192.168.0.1:80",
		"http://192.168.0.1:80",
		"http://2402:4000:2081:3573:e04f:da63:e607:d34d",
		"127.0.0.1",
		"127.0.0.1:8080",
		"192.168.0.1",
		"http://127.0.0.1:8080",
		"2402:4000:2081:3573:e04f:da63:e607:d34d",
		"::1",
		"2001:4860:0:2001::68",
		"[1fff:0:a88:85a3::ac1f]:8001",
		"https://[1fff:0:a88:85a3::ac1f]:8001",
	}
	for _, u := range urls {
		//var host string
		//var port string
		fmt.Println(u)
		val, err1 := url.Parse(u)
		h, p, err2 := net.SplitHostPort(u)
		ip := net.ParseIP(u)

		fmt.Println("url parse===========")
		fmt.Println(err1)
		if err1 == nil {
			fmt.Println("host:	", val.Hostname())
			fmt.Println("port:	", val.Port())
		}
		fmt.Println("split host port==========")
		fmt.Println(err2)
		if err2 == nil {
			fmt.Println("host:	", h)
			fmt.Println("port:	", p)
		}
		fmt.Println("parse ip========")
		fmt.Println("ip:	", ip)
		fmt.Println("\n")
		//if val == nil {
		//	//fmt.Println("not host port	")
		//	ip, portVal, err := net.SplitHostPort(u)
		//	if err != nil {
		//		fmt.Println("Error")
		//	} else {
		//		host = ip
		//		port = portVal
		//	}
		//} else {
		//	host = val.Hostname()
		//	port = val.Port()
		//}

		//fmt.Println("hostname	", host)
		//fmt.Println("port  ", port)
		//fmt.Println()
		//we want: hostname, port

	}
	return true, nil
}

// Endpoint represents the structure of an endpoint.
type DefaultHost struct {
	Host string
	Port string
}

func main() {
	//consulTest()
	//testParList()
	_, _ = urlTest()

}
