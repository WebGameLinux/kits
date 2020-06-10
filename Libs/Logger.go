package Libs

import (
		"encoding/json"
		"fmt"
		"github.com/hashicorp/go-uuid"
		"github.com/sirupsen/logrus"
		"io"
		sysLog "log"
		"os"
		"path/filepath"
		"reflect"
		"sync"
		"time"
)

type LoggerBird struct {
		Channels map[string][]chan interface{}
		Logger             logrus.FieldLogger
		Cache              []map[string]interface{}
		MaxCache           int
		Timeout            time.Duration // 超时
		mut                *sync.Mutex
		Worker             *LoggerBirdWorker
		FlushCachePathRoot string // 保持超时未处理的日志目录
}

type LoggerBirdWorker struct {
		State     int
		CtrChan   chan interface{}
		Mut       *sync.Mutex
		Ctx       *LoggerBird
		WorkCycle time.Duration
		Handler   func(bird *LoggerBird, target ...string)
}

type LevelInterface interface {
		SetLevel(string)
}

const (
		LogTypeError     = "error"
		LogTypeInfo      = "info"
		LogTypeDebug     = "debug"
		LogTypeWarn      = "warn"
		AllNotifyChannel = "*"
		LogId            = "log_id"
		Target           = "target"
		Channel          = "chan"
		LogTexts         = "data"
		TimeAt           = "create_at"
		DefaultCacheSize = 20
		EOL              = "\n"
)

// 工人
func LoggerBirdWorkerNew(handler ...func(args ...interface{}) bool) *LoggerBirdWorker {
		var worker = new(LoggerBirdWorker)
		worker.State = 0
		worker.CtrChan = make(chan interface{}, 2)
		worker.Handler = nil
		worker.Mut = &sync.Mutex{}
		worker.WorkCycle = 60 * time.Second
		if len(handler) > 0 {
				worker.Handler = wrapperWorkerHandler(handler)
		}
		return worker
}

func NewLoggerBird(param ...interface{}) *LoggerBird {
		var bird = new(LoggerBird)
		if len(param) > 0 {
				bird.init(param...)
		}
		if len(bird.Channels) == 0 {
				bird.Channels = map[string][]chan interface{}{}
		}
		if bird.Logger == nil {
				bird.Logger = logrus.StandardLogger()
		}
		if bird.mut == nil {
				bird.mut = &sync.Mutex{}
		}
		if bird.Worker == nil {
				bird.Worker = LoggerBirdWorkerNew()
				bird.Worker.SetCtx(bird)
		}
		if bird.FlushCachePathRoot == "" {
				// os.Args[0]
				path, _ := filepath.Abs(".")
				bird.FlushCachePathRoot = filepath.Dir(path) + string(filepath.Separator) + "logs"
		}
		if bird.MaxCache == 0 {
				bird.MaxCache = DefaultCacheSize
		}
		return bird
}

func (this *LoggerBird) SetLevel(level string) {
		if log, ok := this.Logger.(*logrus.Logger); ok {
				log.SetLevel(levelUint(level))
		}
		if log, ok := this.Logger.(LevelInterface); ok {
				log.SetLevel(level)
		}
}

func (this *LoggerBird) Error(args ...interface{}) {
		this.Send(LogTypeError, args...)
}

func (this *LoggerBird) Debug(args ...interface{}) {
		this.Send(LogTypeDebug, args...)
}

func (this *LoggerBird) Info(args ...interface{}) {
		this.Send(LogTypeInfo, args...)
}

func (this *LoggerBird) Warn(args ...interface{}) {
		this.Send(LogTypeWarn, args...)
}

func (this *LoggerBird) init(args ...interface{}) {
		for _, v := range args {
				if chArr, ok := v.(map[string][]chan interface{}); ok {
						if len(this.Channels) != 0 {
								this.AppendChannel(chArr)
						} else {
								this.Channels = chArr
						}
				}
				if log, ok := v.(logrus.FieldLogger); ok && this.Logger == nil {
						this.Logger = log
				}
				if timeout, ok := v.(time.Duration); ok && this.Timeout == 0 {
						this.Timeout = timeout
				}
				if path, ok := v.(string); ok && this.FlushCachePathRoot == "" {
						state, err := os.Stat(path)
						if err != nil {
								continue
						}
						if state.IsDir() {
								this.FlushCachePathRoot = path
						}
				}
				if worker, ok := v.(*LoggerBirdWorker); ok && this.Worker == nil {
						this.Worker = worker
				}
				if max, ok := v.(int); ok && this.MaxCache == 0 && max > 3 {
						this.MaxCache = max
				}
		}
}

func (this *LoggerBird) AppendChannel(mapper map[string][]chan interface{}) {
		for name, channels := range mapper {
				if name == "" || len(channels) == 0 {
						continue
				}
				if channelSets, ok := this.Channels[name]; ok {
						this.Channels[name] = channelMerge(channelSets, channels)
				} else {
						this.Channels[name] = channelUnique(channels)
				}
		}
}

func (this *LoggerBird) Send(key string, args ...interface{}) {
		this.Notify(key, args)
		if this.Logger != nil {
				args = append(args, key)
				sysLog.Println(args...)
				return
		}
		switch key {
		case LogTypeWarn:
				this.Logger.Warn(args...)

		case LogTypeInfo:
				this.Logger.Info(args...)

		case LogTypeDebug:
				this.Logger.Debug(args...)

		case LogTypeError:
				this.Logger.Error(args...)
		}
}

func (this *LoggerBird) Notify(channel string, args []interface{}) {
		id, _ := uuid.GenerateUUID()
		var msg = map[string]interface{}{
				Channel:  channel,
				Target:   channel,
				LogTexts: args,
				TimeAt:   time.Now().Unix(),
				LogId:    id,
		}
		// 特殊日志处理
		if ch, ok := this.Channels[channel]; ok {
				go this.loop(channel, ch, msg)
		}
		// 所有监听处理
		if ch, ok := this.Channels[AllNotifyChannel]; ok {
				msg[Target] = AllNotifyChannel
				go this.loop(AllNotifyChannel, ch, msg)
		}
}

// 阻塞未处理
func (this *LoggerBird) wait(ch chan interface{}, msg interface{}, target string) {
		if msg == nil {
				return
		}
		var cache map[string]interface{}
		if m, ok := msg.(map[string]interface{}); ok {
				cache = m
		} else {
				id, _ := uuid.GenerateUUID()
				cache = map[string]interface{}{
						Channel: ch, LogTexts:
						msg, Target: target,
						LogId:       id,
				}
		}
		ty := reflect.TypeOf(cache[Channel])
		if ty.Kind() != reflect.Chan && ty.Kind() == reflect.String {
				cache[Channel] = ch
		}
		if len(this.Cache) > this.MaxCache {
				sysLog.Println("logger cache too big , size: ", len(this.Cache))
				this.flush()
		}
		this.Cache = append(this.Cache, cache)
}

// 刷新
func (this *LoggerBird) flush() {
		for _, data := range this.Cache {
				this.save(data)
		}
		this.Cache = this.Cache[0:0]
}

// 保持长期无消耗的日志
func (this *LoggerBird) save(log map[string]interface{}) {
		var file = this.getCacheFile()
		if ch, ok := log[Channel]; ok {
				if reflect.TypeOf(ch).Kind() == reflect.Chan {
						delete(log, Channel)
				}
		}
		logText := this.format(log)
		if logText != "" {
				saveLog(file, []byte(logText), os.ModePerm)
		} else {
				if logText, err := json.Marshal(log); err == nil {
						saveLog(file, logText, os.ModePerm)
				}
		}
		sysLog.Println(log)
}

func (this *LoggerBird) format(log map[string]interface{}) string {
		var (
				at    int64
				level string
				text  string
		)
		if timeAt, ok := log[TimeAt]; ok {
				if at, ok = timeAt.(int64); !ok {
						return ""
				}
		}
		if target, ok := log[Target]; ok {
				if level, ok = target.(string); !ok {
						return ""
				}
		}
		if texts, ok := log[LogTexts]; ok {
				if contents, ok := texts.([]interface{}); ok {
						for _, txt := range contents {
								if str, ok := txt.(string); ok {
										text = text + str
										continue
								}
								if str, ok := txt.(fmt.Stringer); ok {
										text = text + str.String()
								}
						}
				}
		}
		if text == "" {
				return ""
		}
		t := time.Unix(at, 0).Format(time.RFC3339)
		return fmt.Sprintf("[%s] %s %s \n", level, t, text)
}

func (this *LoggerBird) getCacheFile() string {
		y, m, d := time.Now().Date()
		_ = os.MkdirAll(this.FlushCachePathRoot, os.ModePerm)
		file := fmt.Sprintf("log_flush_save.%d-%d-%d.log", y, m, d)
		return this.FlushCachePathRoot + string(filepath.Separator) + file
}

// 检查缓存日志
func (this *LoggerBird) weakUp(name string) {
		if this.Worker == nil {
				this.Worker = LoggerBirdWorkerNew()
				this.Worker.SetCtx(this)
		}
		state := this.Worker.GetState()
		if state == 1 {
				go func() {
						this.Worker.CtrChan <- name
				}()
		}
		if state == 0 {
				this.Worker.Start()
		}
}

func (this *LoggerBird) loop(target string, sets []chan interface{}, data interface{}) {
		this.mut.Lock()
		defer this.mut.Unlock()
		// 唤醒
		this.weakUp(target)
		// 处理当前
		for _, ch := range sets {
				// chan 是否阻塞等待中
				if cap(ch) > len(ch) {
						ch <- data
				} else {
						this.wait(ch, data, target)
				}
		}
}

func (this *LoggerBird) RemoveCache(index int, mapper map[string]interface{}) {
		if len(this.Cache) > index {
				id, _ := mapper[LogId]
				id2, _ := this.Cache[index][LogId]
				if id != id2 {
						this.findAndRemove(id)
				} else {
						this.Cache = MapperArrayPop(index, this.Cache)
				}
		}
}

func (this *LoggerBird) findAndRemove(id interface{}) {
		for i, m := range this.Cache {
				if _id, ok := m[LogId]; ok {
						if _id == id {
								this.Cache = MapperArrayPop(i, this.Cache)
						}
				}
		}
}

func (this *LoggerBirdWorker) SetCtx(ctx *LoggerBird) {
		this.Mut.Lock()
		if this.Ctx == nil {
				this.Ctx = ctx
		}
		this.Mut.Unlock()
}

func (this *LoggerBirdWorker) GetState() int {
		this.Mut.Lock()
		defer this.Mut.Unlock()
		return this.State
}

func (this *LoggerBirdWorker) SetState(state int) {
		this.Mut.Lock()
		defer this.Mut.Unlock()
		this.State = state
}

func (this *LoggerBirdWorker) Start() {
		if this.State == 1 {
				return
		}
		this.Mut.Lock()
		defer this.Mut.Unlock()
		this.Init()
		if this.Handler == nil || this.Ctx == nil {
				return
		}
		go func() {
				this.SetState(1)
				for {
						now := time.Now()
						select {
						// 状态控制
						case state := <-this.CtrChan:
								if n, ok := state.(int); ok {
										if n == -1 {
												goto WorkEnd
										}
								}
								if target, ok := state.(string); ok {
										if target != "" {
												this.SetState(2)
												sysLog.Println("worker....")
												this.Handler(this.Ctx, target)
										}
								}
						// 周期检查
						case <-time.NewTicker(this.WorkCycle).C:
								this.SetState(2)
								sysLog.Println("worker....")
								this.Handler(this.Ctx)
						}
						this.SetState(1)
						end := time.Now()
						sysLog.Println("worker cost : ", end.Sub(now))
				}
		WorkEnd:
				this.SetState(-1)
		}()
}

func (this *LoggerBirdWorker) Init() {
		if this.Handler == nil {
				this.Handler = defaultWorkHandler()
		}
}

// 默认处理器
//  {"chan": chan, "logTexts": msg, "target": target,"log_id":uuid}
func defaultWorkHandler() func(ctx *LoggerBird, target ...string) {
		return func(ctx *LoggerBird, target ...string) {
				var ch string
				if len(target) != 0 && target[0] != "" {
						ch = target[0]
				}

				for i := 0; i < len(ctx.Cache); {
						mapper := ctx.Cache[i]
						if ch != "" {
								if m, ok := mapper[Target]; !ok || m != ch {
										i++
										continue
								}
						}
						if chanSender(mapper) {
								ctx.RemoveCache(i, mapper)
						} else {
								i++
						}
				}
		}
}

// 发送到channel
func chanSender(mapper map[string]interface{}) bool {
		var (
				ok   bool
				v    interface{}
				ch   chan interface{}
				logs []interface{}
		)
		if len(mapper) == 0 {
				return false
		}
		if v, ok = mapper[Channel]; ok {
				if ch, ok = v.(chan interface{}); !ok {
						return false
				}
		}
		if v, ok = mapper[LogTexts]; ok {
				if logs, ok = v.([]interface{}); !ok {
						return false
				}
		}

		if ch != nil && cap(ch) > len(ch) {
				ch <- logs
				return true
		}
		return false
}

// 封装器
func wrapperWorkerHandler(handlers []func(args ...interface{}) bool) func(ctx *LoggerBird, target ...string) {
		return func(ctx *LoggerBird, target ...string) {
				if len(target) == 0 {
						target = append(target, "")
				}
				for _, fn := range handlers {
						if !fn(ctx, target[0]) {
								break
						}
				}
		}
}

// 去重
func channelUnique(arr []chan interface{}) []chan interface{} {
		var newArr []chan interface{}
		for i := 0; i < len(arr); i++ {
				repeat := false
				for j := i + 1; j < len(arr); j++ {
						if arr[i] == arr[j] {
								repeat = true
								break
						}
				}
				if !repeat {
						newArr = append(newArr, arr[i])
				}
		}
		return newArr
}

// 	合并
func channelMerge(array []chan interface{}, add []chan interface{}) []chan interface{} {
		var newArr []chan interface{}
		for _, ch := range add {
				repeat := false
				for _, old := range array {
						if ch == old {
								repeat = true
								break
						}
				}
				if !repeat {
						newArr = append(newArr, ch)
				}
		}
		return newArr
}

func levelUint(name string) logrus.Level {
		for _, level := range logrus.AllLevels {
				if level.String() == name {
						return level
				}
		}
		return logrus.DebugLevel
}

// 移除初对应位置的元素
func MapperArrayPop(index int, array []map[string]interface{}) []map[string]interface{} {
		var num = len(array)
		if num == 0 {
				return array
		}
		if num < index {
				return array
		}
		if index < 0 && -index < num {
				tmp := array[0 : num+index]
				if index == -1 {
						return tmp
				}
				array = append(tmp, array[num+index+1:]...)
				return array
		}
		if index == 0 && num > 1 {
				array = array[1:]
		}
		if index+1 == num {
				array = array[:index]
		}
		if index > 0 && index+1 < num {
				tmp := array[0:index]
				array = append(tmp, array[index+1:]...)
		}
		return array
}

// 保存日志文件
func saveLog(filename string, data []byte, perm os.FileMode) bool {
		var End = []byte(EOL)
		if !matchEnd(End, data) {
				data = append(data, End...)
		}
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
		if err != nil {
				return false
		}
		n, err := f.Write(data)
		if err == nil && n < len(data) {
				err = io.ErrShortWrite
		}
		if err1 := f.Close(); err == nil {
				err = err1
		}
		return err == nil
}

// 是否结尾
func matchEnd(end []byte, data []byte) bool {
		var (
				endStr string
				num    = len(end)
				length = len(data)
				eol    = string(end)
		)
		if num > length {
				return false
		}
		endStr = string(data[length-num:])
		if endStr == eol {
				return true
		}
		return false
}
