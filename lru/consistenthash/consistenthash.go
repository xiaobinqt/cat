package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int
	keys     []int // Stored, hash circle
	// virtual node and real node map, key is virtual node hash, value is real node name
	hashMap map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

// Add function allowing add zero or more real nodes.
// for every real node creating replicas number virtual nodes,
// virtual name is named `strconv.Itoa(i) + key`
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			//log.Printf("consistenthash Add func key: %s, hash is: %d \n", key, hash)
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}

	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// cautions: 这里不太好理解..
	// 如果 idx == len(m.keys)，那顺时针的第一个节点，如果把环拉成直线，就是第一个，说明应选择 m.keys[0]
	// 因为 m.keys 是一个环状结构，可以用取余数的方式来处理这种情况
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
