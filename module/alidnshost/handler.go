/*
 * Copyright (c) 2021 qingchuwudi
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * author bypf2009@vip.qq.com
 * create at 2021/12/10
 */

package alidnshost

import (
	"strings"

	"aliyun-dnshost/module/loger"
	"aliyun-dnshost/module/myip"
	"aliyun-dnshost/module/utils"

	"github.com/alibabacloud-go/domain-20180129/v3/client"
)

// 更新ip
func UpdateDomains(domainService utils.DomainService, broadbandIP map[string]bool, DnsName, IP, IpType string, Domain []*client.QueryDomainListResponseBodyDataDomain) error {

	cacheKey := myip.CacheKey(DnsName, IpType)

	loger.Info("域名总数：[%v]", len(Domain))

	for _, domain := range Domain {
		// 查询阿里云上的记录
		respQueryDnsHost, errQueryDnsHost := domainService.QueryDnsHost(domain.InstanceId)
		if errQueryDnsHost != nil {
			loger.Debug("errQueryDnsHost: %v", errQueryDnsHost)
		}
		for _, dnsHostList := range respQueryDnsHost.Body.DnsHostList {
			for _, ipList := range dnsHostList.IpList {
				// 将记录写入缓存以便比较是否为更新后的ip
				myip.CurrentCache.Put(cacheKey, *ipList, IpType, *dnsHostList.DnsName)
			}
		}
		if !myip.DoesIPChanged(broadbandIP, cacheKey, IP) {
			// 没变化，返回
			loger.Debug("[%s] IP没有发生变化", cacheKey)
			return nil
		}
		// 有变化： 并且缓存中没有有记录
		if myip.CurrentCache.IsNotExist(cacheKey) {

			respCreatingDnsHost, errrespCreatingDnsHost := domainService.CreatingDnsHost(domain.InstanceId, &DnsName, &IP)
			if errrespCreatingDnsHost != nil {
				if ErrorDomainRecordDuplicate(errrespCreatingDnsHost) {
					if respCreatingDnsHost != nil && respCreatingDnsHost.Body != nil {
						// 更新缓存
						myip.CurrentCache.Put(cacheKey, IP, IpType, DnsName)
						loger.Info("添加：已经有记录，更新缓存并跳过添加过程")
					}
					return nil
				}
				return errrespCreatingDnsHost
			}
			respSynchronizingDnsHost, errSynchronizingDnsHost := domainService.SynchronizingDnsHost(domain.InstanceId)
			if errSynchronizingDnsHost != nil {
				loger.Debug("respSynchronizingDnsHost.StatusCode: %v", respSynchronizingDnsHost.StatusCode)
				loger.Debug("errSynchronizingDnsHost: %v", errSynchronizingDnsHost)
			}

			if respCreatingDnsHost != nil && respCreatingDnsHost.Body != nil {
				// 更新缓存
				myip.CurrentCache.Put(cacheKey, IP, IpType, DnsName)
				loger.Info("添加：[%s] 公网IP已添加: %s", cacheKey, IP)
			}
		} else {
			// 有记录，更新
			cache := myip.CurrentCache.GetDnsName(cacheKey)
			loger.Info("[%s]", cache)

			respModifyingDnsHost, errModifyingDnsHost := domainService.ModifyingDnsHost(domain.InstanceId, &DnsName, &IP)
			if errModifyingDnsHost != nil {
				if ErrorDomainRecordDuplicate(errModifyingDnsHost) {
					if respModifyingDnsHost != nil && respModifyingDnsHost.Body != nil {
						// 更新缓存
						myip.CurrentCache.Put(cacheKey, IP, IpType, DnsName)
						loger.Info("更新：已经有相同记录，稍后重试")
					}
					return nil
				}
				return errModifyingDnsHost
			}
			respSynchronizingDnsHost, errSynchronizingDnsHost := domainService.SynchronizingDnsHost(domain.InstanceId)
			if errSynchronizingDnsHost != nil {
				loger.Debug("respSynchronizingDnsHost.StatusCode: %v", respSynchronizingDnsHost.StatusCode)
				loger.Debug("errSynchronizingDnsHost: %v", errSynchronizingDnsHost)
			}
			if respModifyingDnsHost != nil && respModifyingDnsHost.Body != nil {
				// 更新缓存
				myip.CurrentCache.Put(cacheKey, IP, IpType, DnsName)
				loger.Info("[%s] 公网IP更新: %s", cacheKey, IP)
			}
		}
	}
	return nil
}

// 错误处理
func ErrorDomainRecordDuplicate(err error) bool {
	return strings.Contains(err.Error(), "DomainRecordDuplicate")
}
