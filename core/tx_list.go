// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"container/heap"
	"math"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

// nonceHeap is a heap.Interface implementation over 64bit unsigned integers for
// retrieving sorted transactions from the possibly gapped future queue.
// nonceHeap은 64bit unsigned integer환경에서 heap.Interface를 구현한것으로
// 틈이 많이 발생할수있는 future queue에서 정렬된 트렌젝션들을 반환한다
type nonceHeap []uint64

func (h nonceHeap) Len() int           { return len(h) }
func (h nonceHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h nonceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *nonceHeap) Push(x interface{}) {
	*h = append(*h, x.(uint64))
}

func (h *nonceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// txSortedMap is a nonce->transaction hash map with a heap based index to allow
// iterating over the contents in a nonce-incrementing way.
// txSortedMap 함수는 힙 베이스의 색인상의 논스 -> 트렌잭션 해시 맵으로
// 논스 증가 방식으로 컨텐츠를 반복한다
type txSortedMap struct {
	items map[uint64]*types.Transaction // Hash map storing the transaction data
	index *nonceHeap                    // Heap of nonces of all the stored transactions (non-strict mode)
	cache types.Transactions            // Cache of the transactions already sorted
}

// newTxSortedMap creates a new nonce-sorted transaction map.
// newTxSortedMap함수는 논스정렬된 트렌젝션 맵을 반환한다
func newTxSortedMap() *txSortedMap {
	return &txSortedMap{
		items: make(map[uint64]*types.Transaction),
		index: new(nonceHeap),
	}
}

// Get retrieves the current transactions associated with the given nonce.
// Get함수는 주어진 논스에 관련된 현재 트렌젝션을 반환한다
func (m *txSortedMap) Get(nonce uint64) *types.Transaction {
	return m.items[nonce]
}

// Put inserts a new transaction into the map, also updating the map's nonce
// index. If a transaction already exists with the same nonce, it's overwritten.
// Put함수는 새로운 트렌젝션을 맵에 넣고, 동시에 맵의 논스 인덱스를 업데이트 한다
// 만약 동일한 논스를 가진 트렌젝션이 이미 존재한다면 다시 써진다
func (m *txSortedMap) Put(tx *types.Transaction) {
	nonce := tx.Nonce()
	if m.items[nonce] == nil {
		heap.Push(m.index, nonce)
	}
	m.items[nonce], m.cache = tx, nil
}

// Forward removes all transactions from the map with a nonce lower than the
// provided threshold. Every removed transaction is returned for any post-removal
// maintenance.
// Forward함수는 주어진 한계값보다 작은 논스를 가진 모든 트렌젝션을 맵으로 부터 제거한다
// 제거된 모든 트렌젝션은 제거 후 유지를 위해 리턴한다
func (m *txSortedMap) Forward(threshold uint64) types.Transactions {
	var removed types.Transactions

	// Pop off heap items until the threshold is reached
	for m.index.Len() > 0 && (*m.index)[0] < threshold {
		nonce := heap.Pop(m.index).(uint64)
		removed = append(removed, m.items[nonce])
		delete(m.items, nonce)
	}
	// If we had a cached order, shift the front
	if m.cache != nil {
		m.cache = m.cache[len(removed):]
	}
	return removed
}

// Filter iterates over the list of transactions and removes all of them for which
// the specified function evaluates to true.
// 필터 함수는 리스트를 반복하면서 인자로 전달된 함수의 실행결과가 참인 모든 트렌젝션을 제거한다
func (m *txSortedMap) Filter(filter func(*types.Transaction) bool) types.Transactions {
	var removed types.Transactions

	// Collect all the transactions to filter out
	for nonce, tx := range m.items {
		if filter(tx) {
			removed = append(removed, tx)
			delete(m.items, nonce)
		}
	}
	// If transactions were removed, the heap and cache are ruined
	if len(removed) > 0 {
		*m.index = make([]uint64, 0, len(m.items))
		for nonce := range m.items {
			*m.index = append(*m.index, nonce)
		}
		heap.Init(m.index)

		m.cache = nil
	}
	return removed
}

// Cap places a hard limit on the number of items, returning all transactions
// exceeding that limit.
//Cap 함수는 아이템의 최대 한계수를 정하고, 이 한계 숫자를 넘는 모든 트렌젝션을 리턴한다
func (m *txSortedMap) Cap(threshold int) types.Transactions {
	// Short circuit if the number of items is under the limit
	if len(m.items) <= threshold {
		return nil
	}
	// Otherwise gather and drop the highest nonce'd transactions
	var drops types.Transactions

	sort.Sort(*m.index)
	for size := len(m.items); size > threshold; size-- {
		drops = append(drops, m.items[(*m.index)[size-1]])
		delete(m.items, (*m.index)[size-1])
	}
	*m.index = (*m.index)[:threshold]
	heap.Init(m.index)

	// If we had a cache, shift the back
	if m.cache != nil {
		m.cache = m.cache[:len(m.cache)-len(drops)]
	}
	return drops
}

// Remove deletes a transaction from the maintained map, returning whether the
// transaction was found.
// Remove함수는 유지중인 맵에서 트렌젝션을 제거하고 트렌젝션의 발견 여부를 반환한다
func (m *txSortedMap) Remove(nonce uint64) bool {
	// Short circuit if no transaction is present
	_, ok := m.items[nonce]
	if !ok {
		return false
	}
	// Otherwise delete the transaction and fix the heap index
	for i := 0; i < m.index.Len(); i++ {
		if (*m.index)[i] == nonce {
			heap.Remove(m.index, i)
			break
		}
	}
	delete(m.items, nonce)
	m.cache = nil

	return true
}

// Ready retrieves a sequentially increasing list of transactions starting at the
// provided nonce that is ready for processing. The returned transactions will be
// removed from the list.
//
// Note, all transactions with nonces lower than start will also be returned to
// prevent getting into and invalid state. This is not something that should ever
// happen but better to be self correcting than failing!
// Ready함수는 주어진 논스로 부터 실행준비가 된 트렌젝션을의 순차적으로 증가하는 리스트를 반환한다
// 반환된 트렌젝션들을 리스트로 부터 제거될 것이다
// 논스 이하의 모든 트렌젝션 역시 무효상태로 들어가는 것을 방지하기 위해 리턴할 것이다
// 이런 일은 일어나진 않겠지만 실패하는 것보다는 코드로 매니지 해놓는것이 좋다
func (m *txSortedMap) Ready(start uint64) types.Transactions {
	// Short circuit if no transactions are available
	if m.index.Len() == 0 || (*m.index)[0] > start {
		return nil
	}
	// Otherwise start accumulating incremental transactions
	var ready types.Transactions
	for next := (*m.index)[0]; m.index.Len() > 0 && (*m.index)[0] == next; next++ {
		ready = append(ready, m.items[next])
		delete(m.items, next)
		heap.Pop(m.index)
	}
	m.cache = nil

	return ready
}

// Len returns the length of the transaction map.
//Len 함수는 트렌젝션 맵의 길이를 반환한다
func (m *txSortedMap) Len() int {
	return len(m.items)
}

// Flatten creates a nonce-sorted slice of transactions based on the loosely
// sorted internal representation. The result of the sorting is cached in case
// it's requested again before any modifscations are made to the contents.
// Flatten 함수는 내부적 으로 표현되는 완벽하지 않은 정렬 상태에 기반하여
// 논스로 정렬된 트렌젝션의 조각을 반환한다
// 내용에 대한 어떤 수정이 없는 상태에서 또다시 요청된다면 이 정렬 결과는 캐싱된다
func (m *txSortedMap) Flatten() types.Transactions {
	// If the sorting was not cached yet, create and cache it
	if m.cache == nil {
		m.cache = make(types.Transactions, 0, len(m.items))
		for _, tx := range m.items {
			m.cache = append(m.cache, tx)
		}
		sort.Sort(types.TxByNonce(m.cache))
	}
	// Copy the cache to prevent accidental modifications
	txs := make(types.Transactions, len(m.cache))
	copy(txs, m.cache)
	return txs
}

// txList is a "list" of transactions belonging to an account, sorted by account
// nonce. The same type can be used both for storing contiguous transactions for
// the executable/pending queue; and for storing gapped transactions for the non-
// executable/future queue, with minor behavioral changes.
// txList는 하나의 어카운트에 속하는 트렌젝션의 리스트이며, 어카운트 논스에 의해 정렬된다.
// 펜딩큐와 실행큐양쪽에서 연속된 트렌젝션을 저장하기 위해 사용된다
type txList struct {
	strict bool         // Whether nonces are strictly continuous or not
	txs    *txSortedMap // Heap indexed sorted hash map of the transactions

	costcap *big.Int // Price of the highest costing transaction (reset only if exceeds balance)
	gascap  uint64   // Gas limit of the highest spending transaction (reset only if exceeds block limit)
}

// newTxList create a new transaction list for maintaining nonce-indexable fast,
// gapped, sortable transaction lists.
// newTxList함수는 논스 정렬된 빠르고, 듬성듬성하고 정렬가능한 트렌젝션 리스트를 유지하기 위해
// 새로운 트렌젝션을 생성한다
func newTxList(strict bool) *txList {
	return &txList{
		strict:  strict,
		txs:     newTxSortedMap(),
		costcap: new(big.Int),
	}
}

// Overlaps returns whether the transaction specified has the same nonce as one
// already contained within the list.
// Overlaps 함수는 지정된 트렌젝션이 이미 리스트에 포함된 논스와 
// 같은지 여부를 반환한다
func (l *txList) Overlaps(tx *types.Transaction) bool {
	return l.txs.Get(tx.Nonce()) != nil
}

// Add tries to insert a new transaction into the list, returning whether the
// transaction was accepted, and if yes, any previous transaction it replaced.
//
// If the new transaction is accepted into the list, the lists' cost and gas
// thresholds are also potentially updated.
// Add함수는 트렌젝션을 리스트로 넣는 시도를 하며 트렌젝션이 수락되었는지 여부를 반환한다
// 만약 수락되었다면 기존 트렌젝션은 교체된다
// 만약 새로운 트렌젝션이 리스트에 수락되었다면 리스트의 비용과 가스 한도역시 갱신될 수 있다
func (l *txList) Add(tx *types.Transaction, priceBump uint64) (bool, *types.Transaction) {
	// If there's an older better transaction, abort
	old := l.txs.Get(tx.Nonce())
	if old != nil {
		threshold := new(big.Int).Div(new(big.Int).Mul(old.GasPrice(), big.NewInt(100+int64(priceBump))), big.NewInt(100))
		// Have to ensure that the new gas price is higher than the old gas
		// price as well as checking the percentage threshold to ensure that
		// this is accurate for low (Wei-level) gas price replacements
		if old.GasPrice().Cmp(tx.GasPrice()) >= 0 || threshold.Cmp(tx.GasPrice()) > 0 {
			return false, nil
		}
	}
	// Otherwise overwrite the old transaction with the current one
	l.txs.Put(tx)
	if cost := tx.Cost(); l.costcap.Cmp(cost) < 0 {
		l.costcap = cost
	}
	if gas := tx.Gas(); l.gascap < gas {
		l.gascap = gas
	}
	return true, old
}

// Forward removes all transactions from the list with a nonce lower than the
// provided threshold. Every removed transaction is returned for any post-removal
// maintenance.
// Forward함수는 주어진 한계값보다 작은 논스를 가진 모든 트렌젝션을 맵으로 부터 제거한다
// 제거된 모든 트렌젝션은 제거 후 유지를 위해 리턴한다
func (l *txList) Forward(threshold uint64) types.Transactions {
	return l.txs.Forward(threshold)
}

// Filter removes all transactions from the list with a cost or gas limit higher
// than the provided thresholds. Every removed transaction is returned for any
// post-removal maintenance. Strict-mode invalidated transactions are also
// returned.
//
// This method uses the cached costcap and gascap to quickly decide if there's even
// a point in calculating all the costs or if the balance covers all. If the threshold
// is lower than the costgas cap, the caps will be reset to a new high after removing
// the newly invalidated transactions.
// Filter함수는 비용이나 가스 한도가 주어진 한도보다 높을 경우 모든 트렌젝션을 리스트로 부터 제거한다
// 모든 제거된 트렌젝션은 제거후처리를 유지하기 위해 반환된다
// 엄격하게 검증된 트렌젝션도 역시 반환된다
// 이 함수는 모든 비용이나 잔고등을 빠르게 계산하기 위해 캐싱된 비용한도와 가스 한도를 사용한다
// 만약 한도가 가스 비용한도보다 작다면 한도가 새롭게 무효화된 트렌젝션들을 제거한 후
// 더높은 값으로 새로 설정될 것이다
func (l *txList) Filter(costLimit *big.Int, gasLimit uint64) (types.Transactions, types.Transactions) {
	// If all transactions are below the threshold, short circuit
	if l.costcap.Cmp(costLimit) <= 0 && l.gascap <= gasLimit {
		return nil, nil
	}
	l.costcap = new(big.Int).Set(costLimit) // Lower the caps to the thresholds
	l.gascap = gasLimit

	// Filter out all the transactions above the account's funds
	removed := l.txs.Filter(func(tx *types.Transaction) bool { return tx.Cost().Cmp(costLimit) > 0 || tx.Gas() > gasLimit })

	// If the list was strict, filter anything above the lowest nonce
	var invalids types.Transactions

	if l.strict && len(removed) > 0 {
		lowest := uint64(math.MaxUint64)
		for _, tx := range removed {
			if nonce := tx.Nonce(); lowest > nonce {
				lowest = nonce
			}
		}
		invalids = l.txs.Filter(func(tx *types.Transaction) bool { return tx.Nonce() > lowest })
	}
	return removed, invalids
}

// Cap places a hard limit on the number of items, returning all transactions
// exceeding that limit.
// itrem 수량의 한도를 설정하고 초과하는 모든 트렌젠션을 반환해버린다
func (l *txList) Cap(threshold int) types.Transactions {
	return l.txs.Cap(threshold)
}

// Remove deletes a transaction from the maintained list, returning whether the
// transaction was found, and also returning any transaction invalidated due to
// the deletion (strict mode only).
// Remove함수는 관리중인 리스트로부터 트렌젝션을 제거하며
// 트렌젝션이 찾아졌거나 삭제때문에 무효화된 모든 트렌젝션을 반환한다
func (l *txList) Remove(tx *types.Transaction) (bool, types.Transactions) {
	// Remove the transaction from the set
	nonce := tx.Nonce()
	if removed := l.txs.Remove(nonce); !removed {
		return false, nil
	}
	// In strict mode, filter out non-executable transactions
	if l.strict {
		return true, l.txs.Filter(func(tx *types.Transaction) bool { return tx.Nonce() > nonce })
	}
	return true, nil
}

// Ready retrieves a sequentially increasing list of transactions starting at the
// provided nonce that is ready for processing. The returned transactions will be
// removed from the list.
//
// Note, all transactions with nonces lower than start will also be returned to
// prevent getting into and invalid state. This is not something that should ever
// happen but better to be self correcting than failing!
// Ready함수는 처리 가능한 순차적으로 증가하는 트렌젝션의 리스트를 주어진 논스로 부터 반환한다
// 반환된 트렌젝션은 리스트로 부터 삭제된다
// 주어진 논스보다 낮은 모든 트렌젝션 역시 무효 상태와 리스트의 사용방지를 위해 반환된다

func (l *txList) Ready(start uint64) types.Transactions {
	return l.txs.Ready(start)
}

// Len returns the length of the transaction list.
func (l *txList) Len() int {
	return l.txs.Len()
}

// Empty returns whether the list of transactions is empty or not.
func (l *txList) Empty() bool {
	return l.Len() == 0
}

// Flatten creates a nonce-sorted slice of transactions based on the loosely
// sorted internal representation. The result of the sorting is cached in case
// it's requested again before any modifications are made to the contents.
// Flatten 함수는 내부적 으로 표현되는 완벽하지 않은 정렬 상태에 기반하여
// 논스로 정렬된 트렌젝션의 조각을 반환한다
// 내용에 대한 어떤 수정이 없는 상태에서 또다시 요청된다면 이 정렬 결과는 캐싱된다
func (l *txList) Flatten() types.Transactions {
	return l.txs.Flatten()
}

// priceHeap is a heap.Interface implementation over transactions for retrieving
// price-sorted transactions to discard when the pool fills up.
// priceHeap은 트렌젝션들을 위한 heap 인터페이스를 구현하며
// 풀이 가득 찼을때 버릴 가격 정렬된 트렌젝션을 반환한다
type priceHeap []*types.Transaction

func (h priceHeap) Len() int      { return len(h) }
func (h priceHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h priceHeap) Less(i, j int) bool {
	// Sort primarily by price, returning the cheaper one
	switch h[i].GasPrice().Cmp(h[j].GasPrice()) {
	case -1:
		return true
	case 1:
		return false
	}
	// If the prices match, stabilize via nonces (high nonce is worse)
	return h[i].Nonce() > h[j].Nonce()
}

func (h *priceHeap) Push(x interface{}) {
	*h = append(*h, x.(*types.Transaction))
}

func (h *priceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// txPricedList is a price-sorted heap to allow operating on transactions pool
// contents in a price-incrementing way.
// txPricedList는 트렌젝션 풀의 내용이 가격이 증가하는 형태로 동작하게 하기 위한 가격 정렬된 힙이다
type txPricedList struct {
	all    *map[common.Hash]*types.Transaction // Pointer to the map of all transactions
	items  *priceHeap                          // Heap of prices of all the stored transactions
	stales int                                 // Number of stale price points to (re-heap trigger)
}

// newTxPricedList creates a new price-sorted transaction heap.
// newTxPricedList함수는 새롭게 가격정렬된 트렌젝션 힙을 생성한다
func newTxPricedList(all *map[common.Hash]*types.Transaction) *txPricedList {
	return &txPricedList{
		all:   all,
		items: new(priceHeap),
	}
}

// Put inserts a new transaction into the heap.
// Put함수는 새로운 트렌젝션을 힙에 삽입한다
func (l *txPricedList) Put(tx *types.Transaction) {
	heap.Push(l.items, tx)
}

// Removed notifies the prices transaction list that an old transaction dropped
// from the pool. The list will just keep a counter of stale objects and update
// the heap if a large enough ratio of transactions go stale.
// Removed함수는 풀로부터 제거되는 오래된 트렌젝션의 가격 리스트를 알린다
// 이 리스트는 상태 오브젝트의 숫자만을 저장하며
// 트렌젝션이 떨어질때의 충분히 큰 비율로 힙을 업데이트 한다
func (l *txPricedList) Removed() {
	// Bump the stale counter, but exit if still too low (< 25%)
	l.stales++
	if l.stales <= len(*l.items)/4 {
		return
	}
	// Seems we've reached a critical number of stale transactions, reheap
	reheap := make(priceHeap, 0, len(*l.all))

	l.stales, l.items = 0, &reheap
	for _, tx := range *l.all {
		*l.items = append(*l.items, tx)
	}
	heap.Init(l.items)
}

// Cap finds all the transactions below the given price threshold, drops them
// from the priced list and returs them for further removal from the entire pool.
// Cap 함수는 주어진 가격보다 높은 모든 트렌젝션들을 찾아 가격 리스트에서 제거하고
// 전체 풀로부터 제거하기 위해 반환한다

func (l *txPricedList) Cap(threshold *big.Int, local *accountSet) types.Transactions {
	drop := make(types.Transactions, 0, 128) // Remote underpriced transactions to drop
	save := make(types.Transactions, 0, 64)  // Local underpriced transactions to keep

	for len(*l.items) > 0 {
		// Discard stale transactions if found during cleanup
		tx := heap.Pop(l.items).(*types.Transaction)
		if _, ok := (*l.all)[tx.Hash()]; !ok {
			l.stales--
			continue
		}
		// Stop the discards if we've reached the threshold
		if tx.GasPrice().Cmp(threshold) >= 0 {
			save = append(save, tx)
			break
		}
		// Non stale transaction found, discard unless local
		if local.containsTx(tx) {
			save = append(save, tx)
		} else {
			drop = append(drop, tx)
		}
	}
	for _, tx := range save {
		heap.Push(l.items, tx)
	}
	return drop
}

// Underpriced checks whether a transaction is cheaper than (or as cheap as) the
// lowest priced transaction currently being tracked.
// Underpriced 함수는 트렌젝션이 현재 추적중인 가장 저가의 트렌젝션보다 싸거나 같은지 확인한다
func (l *txPricedList) Underpriced(tx *types.Transaction, local *accountSet) bool {
	// Local transactions cannot be underpriced
	if local.containsTx(tx) {
		return false
	}
	// Discard stale price points if found at the heap start
	for len(*l.items) > 0 {
		head := []*types.Transaction(*l.items)[0]
		if _, ok := (*l.all)[head.Hash()]; !ok {
			l.stales--
			heap.Pop(l.items)
			continue
		}
		break
	}
	// Check if the transaction is underpriced or not
	if len(*l.items) == 0 {
		log.Error("Pricing query for empty pool") // This cannot happen, print to catch programming errors
		return false
	}
	cheapest := []*types.Transaction(*l.items)[0]
	return cheapest.GasPrice().Cmp(tx.GasPrice()) >= 0
}

// Discard finds a number of most underpriced transactions, removes them from the
// priced list and returns them for further removal from the entire pool.
// Discard 함수는 가장 저렴한 트렌젝션의 갯수를 찾아서 가격 리스트에서 제거하고
// 전체 풀에서 제거하기 위해 반환한다
func (l *txPricedList) Discard(count int, local *accountSet) types.Transactions {
	drop := make(types.Transactions, 0, count) // Remote underpriced transactions to drop
	save := make(types.Transactions, 0, 64)    // Local underpriced transactions to keep

	for len(*l.items) > 0 && count > 0 {
		// Discard stale transactions if found during cleanup
		tx := heap.Pop(l.items).(*types.Transaction)
		if _, ok := (*l.all)[tx.Hash()]; !ok {
			l.stales--
			continue
		}
		// Non stale transaction found, discard unless local
		if local.containsTx(tx) {
			save = append(save, tx)
		} else {
			drop = append(drop, tx)
			count--
		}
	}
	for _, tx := range save {
		heap.Push(l.items, tx)
	}
	return drop
}
