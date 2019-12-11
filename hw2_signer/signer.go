package main

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type workersChanel chan interface{}
type sliceOfWorkersChanels []workersChanel

var simMd5 uint32

func md5Wraper(data string) string {
	for {
		if swapped := atomic.CompareAndSwapUint32(&simMd5, 0, 1); !swapped {
			time.Sleep(time.Millisecond)
		} else {
			result := DataSignerMd5(data)
			atomic.StoreUint32(&simMd5, 0)
			return result
		}
	}
}

/*
ExecutePipeline ...
*/
func ExecutePipeline(jobs ...job) {
	var jobsCount = len(jobs)
	var workersChanels = make(sliceOfWorkersChanels, jobsCount+1, jobsCount+1)
	wg := &sync.WaitGroup{}
	workersChanels[0] = make(workersChanel, 1)
	close(workersChanels[0])

	for jobI, jobItem := range jobs {
		wg.Add(1)
		workersChanels[jobI+1] = make(workersChanel, 1)
		go func(funcJob job, in, out chan interface{}) {
			funcJob(in, out)
			close(out)
			wg.Done()
		}(jobItem, workersChanels[jobI], workersChanels[jobI+1])
		runtime.Gosched()
	}
	wg.Wait()

}

/*
SingleHash ...
*/
var SingleHash job = func(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for dataRaw := range in {
		data := strconv.Itoa(dataRaw.(int))
		//fmt.Println("SingleHash data ", data)
		/*data, ok := dataRaw.(string)

		if !ok {
			fmt.Println("SingleHash error: cant convert result data to string")
		}*/
		wg.Add(1)

		go func(data string, outCh chan interface{}) {
			var crc32, md5, crc32md5 string
			var result string
			wg2 := &sync.WaitGroup{}
			wg2.Add(2)

			go func() {
				crc32 = DataSignerCrc32(data)
				wg2.Done()
			}()

			go func() {
				md5 = md5Wraper(data)
				crc32md5 = DataSignerCrc32(md5)
				wg2.Done()
			}()
			runtime.Gosched()
			wg2.Wait()

			result = crc32 + "~" + crc32md5
			outCh <- result
			wg.Done()
		}(data, out)
		/*fmt.Println("SingleHash md5(data) ", md5)
		fmt.Println("SingleHash crc32(md5(data))", crc32md5)
		fmt.Println("SingleHash crc32(data) ", crc32)
		fmt.Println("SingleHash result ", result)*/

	}
	runtime.Gosched()
	wg.Wait()

}

/*
MultiHash ...
*/
var MultiHash job = func(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for dataRaw := range in {
		data, ok := dataRaw.(string)

		if !ok {
			fmt.Println("MultiHash error: cant convert result data to string")
		}
		//fmt.Println("MultiHash - data:", data)
		// fmt.Println("MultiHash calc result ok: ", ok)
		wg.Add(1)

		go func(data string, outCh chan interface{}) {
			var result string
			var results = make([]string, 7, 7)
			wg2 := &sync.WaitGroup{}

			for i := 0; i < 6; i++ {
				wg2.Add(1)

				go func(res *string, n int) {
					crc32 := DataSignerCrc32(strconv.Itoa(n) + data)
					//fmt.Printf("MultiHash - step%d:%s\n", i, crc32)
					*res = crc32
					wg2.Done()
				}(&results[i], i)
			}
			runtime.Gosched()
			wg2.Wait()
			result = strings.Join(results, "")
			/*for i := 0; i < 6; i++ {
				result += results[i]
			}*/
			//fmt.Println("MultiHash result ", result)
			// fmt.Println("SingleHash result = ", result)
			outCh <- result
			wg.Done()
		}(data, out)
	}
	runtime.Gosched()
	wg.Wait()
	// fmt.Printf("\n\nMultiHash results:\n%#v\n\n", results)
}

/*
CombineResults ...
*/
var CombineResults job = func(in, out chan interface{}) {
	var inputArr = make([]string, 0, 10)
	var result string

	for input := range in {
		inputArr = append(inputArr, input.(string))
	}
	sort.Strings(inputArr)
	result = strings.Join(inputArr, "_")
	out <- result
	// fmt.Println("CombineResults result ", result)
}

func main() {

	/*var ok = true
	var recieved uint32
	freeFlowJobs := []job{
		job(func(in, out chan interface{}) {
			fmt.Println("func1")
			out <- 1
			fmt.Println("func1:sended 1")
			time.Sleep(10 * time.Millisecond)
			currRecieved := atomic.LoadUint32(&recieved)
			// в чем тут суть
			// если вы накапливаете значения, то пока вся функция не отрабоатет - дальше они не пойдут
			// тут я проверяю, что счетчик увеличился в следующей функции
			// это значит что туда дошло значение прежде чем текущая функция отработала
			if currRecieved == 0 {
				ok = false
			}
			fmt.Println("func1:end")
		}),
		job(func(in, out chan interface{}) {
			fmt.Println("func2")
			for _ = range in {
				atomic.AddUint32(&recieved, 1)
				fmt.Println("func2:", recieved)
			}
			fmt.Println("func2:end")
		}),
	}
	ExecutePipeline(freeFlowJobs...)

	if !ok || recieved == 0 {
		fmt.Println("no value free flow - dont collect them")
	}*/

	/*in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)
	in <- "qwerty12345"
	SingleHash(in, out)
	res := <-out

	fmt.Println(res)*/

	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	//inputData := []int{0, 1}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				fmt.Println("main error: cant convert result data to string")
			}
			testResult := data
			fmt.Println(testResult)
		}),
	}

	//start := time.Now()

	ExecutePipeline(hashSignJobs...)

}
