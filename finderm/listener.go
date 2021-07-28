package finderm

import (
	"sync"
)

type queues struct {
	cs map[int64]chan interface{}
	l sync.Mutex
}

func newQueue()*queues{
	return &queues{
		cs: map[int64]chan interface{}{},
	}
}

func (q *queues)Send(e interface{}){
	q.l.Lock()
	defer q.l.Unlock()
	for _, c := range q.cs {
	bb:
		for{
			select {
			case c<-e:
				break bb
			default:
				dropOne(c)
			}
		}
	}
}

func dropOne(c chan interface{}){
	select {
	case <-c:
	default:

	}
}

func (q *queues)Receive(channel int64)interface{}{
	q.l.Lock()
	cc:=q.cs[channel]
	if cc == nil{
		cc = make(chan interface{},20)
		q.cs[channel] = cc
	}
	q.l.Unlock()
	res:=<-cc
	return res

}


type listener struct {
	listeners map[string]*queues
	lock sync.Mutex
}

func newListener()*listener{
	return &listener{
		listeners: map[string]*queues{},
		lock:      sync.Mutex{},
	}
}
// 1，2，3，4
func (l *listener) Listen(key string, queue int64) (interface{}, error) {
	l.lock.Lock()
	lss:=l.listeners[key]
	if lss == nil{
		lss = newQueue()
		l.listeners[key] = lss
	}
	l.lock.Unlock()
	res:=lss.Receive(queue)
	return res, nil
}

func (l *listener) Send(key string, e interface{}) error {
	l.lock.Lock()
	lss:=l.listeners[key]
	l.lock.Unlock()
	if lss == nil{
		return nil
	}
	lss.Send(e)
	return nil
	//return fmt.Errorf("%s,event chan is full",key)
}

var (
	serviceListener = newListener()
	configListener  = newListener()
)
