// Copyright Ngo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package redis

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/NetEase-Media/ngo/pkg/adapter/log"
	"github.com/go-redis/redis/v8"
)

func NewShardedSentinelClient(opt *Options) *redisContainer {
	sentinels := make(map[string]*redis.SentinelClient, len(opt.Addr))
	masterAddrs := make(map[string]string, len(opt.Addr))
	ssc := &ShardedSentinelClient{
		opt:         opt,
		sentinels:   sentinels,
		masterNames: opt.MasterNames,
		masterAddrs: masterAddrs,
	}

	ctx := context.Background()
	for i := range opt.Addr {
		sentinels[opt.Addr[i]] = redis.NewSentinelClient(sentinelOptions(opt, opt.Addr[i]))
	}

	sis := make([]*ShardInfo, 0, len(opt.MasterNames))
	for _, name := range opt.MasterNames {
		var addr string
		for i := range opt.Addr {
			sentinel := sentinels[opt.Addr[i]]
			masterAddr, err := sentinel.GetMasterAddrByName(ctx, name).Result()
			if err != nil {
				log.Errorf("sentinel: GetMasterAddrByName master=%s failed: %s",
					name, err)
				continue
			}
			addr = net.JoinHostPort(masterAddr[0], masterAddr[1])
			masterAddrs[name] = addr
			break
		}

		if len(addr) == 0 {
			panic(fmt.Sprintf("sentinel: GetMasterAddrByName master=%s all failed", name))
		}
		tmp := clientOptions(opt, addr)
		shardName := name
		// 兼容旧分片名称规则，避免线上rehash
		if opt.AutoGenShardName {
			shardName = ""
		}
		sis = append(sis, &ShardInfo{
			id:     name,
			name:   shardName,
			client: NewClient(tmp),
			weight: 1,
		})
	}

	baseClient := NewShardedClient(sis)
	c := &redisContainer{
		Redis:     baseClient,
		opt:       *opt,
		redisType: RedisTypeShardedSentinel,
	}
	ssc.c = c
	go ssc.listen(ctx)
	return c
}

type ShardedSentinelClient struct {
	opt         *Options
	sentinels   map[string]*redis.SentinelClient
	masterNames []string
	masterAddrs map[string]string
	sync.Mutex

	c *redisContainer
}

func (ssc *ShardedSentinelClient) listen(ctx context.Context) {
	for k, v := range ssc.sentinels {
		if v == nil {
			ssc.sentinels[k] = redis.NewSentinelClient(sentinelOptions(ssc.opt, k))
		}
		pubsub := ssc.sentinels[k].Subscribe(ctx, "+switch-master")
		go func(pubsub *redis.PubSub) {
			ch := pubsub.Channel()
			for msg := range ch {
				if msg.Channel == "+switch-master" {
					parts := strings.Split(msg.Payload, " ")
					masterName := parts[0]
					if _, exists := Find(ssc.masterNames, masterName); !exists {
						log.Warnf("sentinel: ignore addr for master=%q", parts[0])
						continue
					}
					addr := net.JoinHostPort(parts[3], parts[4])

					ssc.Lock()
					if ssc.masterAddrs[masterName] == addr {
						log.Warnf("sentinel: addr for master=%q is not change", parts[0])
						ssc.Unlock()
						continue
					}

					client := ssc.c.Redis.(*ShardedClient)
					shardName := masterName
					// 兼容旧分片名称规则，避免线上rehash
					if ssc.opt.AutoGenShardName {
						shardName = ""
					}
					client.ChangeShardInfo(masterName, &ShardInfo{
						id:     masterName,
						name:   shardName,
						client: NewClient(clientOptions(ssc.opt, addr)),
						weight: 1,
					})
					log.Infof("sentinel: switch master \"%v\"", msg.Payload)
					ssc.masterAddrs[masterName] = addr
					ssc.Unlock()
				}
			}
		}(pubsub)
	}
}

func sentinelOptions(opt *Options, addr string) *redis.Options {
	return &redis.Options{
		Addr:               addr,
		DB:                 0,
		MaxRetries:         opt.MaxRetries,
		MinRetryBackoff:    opt.MinRetryBackoff,
		MaxRetryBackoff:    opt.MaxRetryBackoff,
		DialTimeout:        opt.DialTimeout,
		ReadTimeout:        opt.ReadTimeout,
		WriteTimeout:       opt.WriteTimeout,
		PoolSize:           opt.PoolSize,
		PoolTimeout:        opt.PoolTimeout,
		IdleTimeout:        opt.IdleTimeout,
		IdleCheckFrequency: opt.IdleCheckFrequency,
		MinIdleConns:       opt.MinIdleConns,
		MaxConnAge:         opt.MaxConnAge,
		TLSConfig:          opt.TLSConfig,
	}
}

func clientOptions(opt *Options, addr string) *Options {
	return &Options{
		Addr:               []string{addr},
		DB:                 0,
		Password:           opt.Password,
		MaxRetries:         opt.MaxRetries,
		MinRetryBackoff:    opt.MinRetryBackoff,
		MaxRetryBackoff:    opt.MaxRetryBackoff,
		DialTimeout:        opt.DialTimeout,
		ReadTimeout:        opt.ReadTimeout,
		WriteTimeout:       opt.WriteTimeout,
		PoolSize:           opt.PoolSize,
		PoolTimeout:        opt.PoolTimeout,
		IdleTimeout:        opt.IdleTimeout,
		IdleCheckFrequency: opt.IdleCheckFrequency,
		MinIdleConns:       opt.MinIdleConns,
		MaxConnAge:         opt.MaxConnAge,
		TLSConfig:          opt.TLSConfig,
	}
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
