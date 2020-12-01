package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"math/rand"
	"os"
	serviceDiscovery "serviceDiscovery/consul"
	"strconv"
	"strings"
	"sync"
	"time"
)

func consulTest() {
	cl, _ := api.NewClient(api.DefaultConfig())
	client, _ := serviceDiscovery.NewConsulClient(cl)
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

	h := cl.Health()
	hc, _, _ := h.Service(serviceName, tags[0], true, &qo)
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
		case <-time.After(12 * time.Second):
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
func parseList(str string) []string {
	s := strings.Split(str, ",")
	for i, _ := range s {
		s[i] = strings.ReplaceAll(s[i], "[", "")
		s[i] = strings.ReplaceAll(s[i], "]", "")
		if s[i] == "*" {
			s[i] = ""
		}
	}
	return s
}
func parseSyntax(value string) (*serviceDiscovery.QueryString, error) {
	//[dc1,dc2].namespace.serviceA.[tag1,tag2]
	split := strings.Split(value, ":")
	if len(split) != 2 {
		return nil, errors.New("bad query syntax")
	}
	str := strings.Split(split[1], ".")
	qCategory := len(str)
	if qCategory == 3 { //datacenters, service name, tags
		queryString := serviceDiscovery.QueryString{
			Datacenters: parseList(str[0]),
			ServiceName: str[1],
			Namespace:   "",
			Tags:        parseList(str[2]),
		}
		return &queryString, nil
	} else if qCategory == 4 { //datacenters, namespace, service name, tags
		queryString := serviceDiscovery.QueryString{
			Datacenters: parseList(str[0]),
			ServiceName: str[2],
			Namespace:   str[1],
			Tags:        parseList(str[2]),
		}
		return &queryString, nil
	}
	return nil, errors.New("bad query syntax")

}

func testParseSyntax() {
	l := []string{"consul:[dc1,dc2].namespace.serviceA.[tag1,tag2]",
		"[dc1,dc2].namespace.serviceA.[tag1,tag2]",
		"consul:[dc1,dc2.fdr].namespace.serviceA.[tag1,tag2]",
		"",
	}
	for _, i := range l {
		a, e := parseSyntax(i)
		fmt.Println(a, e)
	}
}
func testParseList() {
	s := []string{"[dc1,dc2,aws-us-central-1]", "[]", "", "[*]"}
	for _, i := range s {
		fmt.Println(parseList(i))
	}
}

func main() {
	//consulTest()
	//testParseList()
	testParseSyntax()
}
