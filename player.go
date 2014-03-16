package bombertcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/aybabtme/bomberman/logger"
	"github.com/aybabtme/bomberman/player"
	"net"
	"strings"
	"sync"
)

type TcpPlayer struct {
	stateL sync.RWMutex
	state  player.State
	l      *logger.Logger
	// Comms
	update  chan player.State
	outMove chan player.Move
}

func NewTcpPlayer(state player.State, laddr string, log *logger.Logger) player.Player {
	t := &TcpPlayer{
		l:       log,
		stateL:  sync.RWMutex{},
		state:   state,
		update:  make(chan player.State, 10), // Buffer some responses, if network is slow
		outMove: make(chan player.Move, 1),   // Rate-limiting to 1 move per turn
	}

	go func() {
		addr, err := net.ResolveTCPAddr("tcp", laddr)
		if err != nil {
			log.Errorf("resolving laddr, %v", err)
			return
		}

		if err := t.listenForPlayer(addr); err != nil {
			log.Errorf("[TCP] connecting to player, %v", err)
		}
	}()

	return t
}

func (t *TcpPlayer) listenForPlayer(addr *net.TCPAddr) (err error) {
	var l *net.TCPListener
	l, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return fmt.Errorf("listening on addr, %v", err)
	}
	defer func() { err = l.Close() }()

	alive := true

	for alive {
		err := t.acceptPlayerConn(l)
		if err != nil {
			return err
		}
		t.stateL.RLock()
		alive = t.state.Alive
		t.stateL.RUnlock()
	}

	return nil
}

func (t *TcpPlayer) acceptPlayerConn(l *net.TCPListener) (err error) {
	var conn *net.TCPConn
	conn, err = l.AcceptTCP()
	if err != nil {
		return fmt.Errorf("accepting, %v", err)
	}
	defer func() { err = conn.Close() }()

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() { t.sendUpdates(w, wg) }()
	go func() { t.receiveMoves(r, wg) }()
	wg.Wait()

	return nil
}

func (t *TcpPlayer) sendUpdates(w *bufio.Writer, wg *sync.WaitGroup) {
	defer wg.Done()

	enc := json.NewEncoder(w)

	lastTurnSent := -1
	for update := range t.update {
		t.stateL.Lock()
		t.state = update
		t.stateL.Unlock()

		if update.Turn == lastTurnSent {
			continue
		}

		if err := enc.Encode(update); err != nil {
			t.l.Errorf("[TCP] sending update to player, %v", err)
			return
		}

		if _, err := w.WriteRune('\n'); err != nil {
			t.l.Errorf("[TCP] sending EOL to player, %v", err)
			return
		}

		if err := w.Flush(); err != nil {
			t.l.Errorf("[TCP] flushing update to player, %v", err)
			return
		}
		lastTurnSent = update.Turn

		if !update.Alive {
			return
		}
	}
}

func (t *TcpPlayer) receiveMoves(r *bufio.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	t.stateL.RLock()
	alive := t.state.Alive
	t.stateL.RUnlock()

	for alive {

		moveStr, err := r.ReadString('\n')
		if err != nil {
			t.l.Errorf("[TCP] reading move from connection, %v", err)
			return
		}
		var m player.Move
		switch player.Move(strings.TrimSpace(moveStr)) {
		case player.Up:
			m = player.Up
		case player.Down:
			m = player.Down
		case player.Left:
			m = player.Left
		case player.Right:
			m = player.Right
		case player.PutBomb:
			m = player.PutBomb
		default:
			t.l.Errorf("[TCP] invalid move string")
			t.l.Debugf("[TCP] move='%s'", moveStr)
			continue
		}
		t.outMove <- m

		t.stateL.RLock()
		alive = t.state.Alive
		t.stateL.RUnlock()
	}
}

func (t *TcpPlayer) Name() string {
	t.stateL.RLock()
	defer t.stateL.RUnlock()
	name := t.state.Name
	return name
}

func (t *TcpPlayer) Move() <-chan player.Move {
	return t.outMove
}

func (t *TcpPlayer) Update() chan<- player.State {
	return t.update
}
