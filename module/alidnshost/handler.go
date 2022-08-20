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

	"github.com/alibabacloud-go/domain-20180129/v3/client"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
)

// 更新ip
func UpdateDomains(cli client.Client, broadbandIP map[string]bool, DnsName, IP, ipType string) error {
	runtime := &util.RuntimeOptions{}
	cacheKey := myip.CacheKey(DnsName, ipType)
	if !myip.DoesIPChanged(broadbandIP, cacheKey, IP) {
		// 没变化，返回
		loger.Debug("[%s] IP没有发生变化", cacheKey)
		return nil
	}
	respQueryDomain, errQueryDomain := QueryDomain(cli, runtime)
	if errQueryDomain != nil {
		loger.Debug("errQueryDomain: %v", errQueryDomain)
	}
	loger.Info("域名总数：[%v]", len(respQueryDomain.Body.Data.Domain))
	//resp.Body.Data.Domain[0].InstanceId

	for _, domain := range respQueryDomain.Body.Data.Domain {
		respQuery, errQueryDnsHost := QueryDnsHost(cli, runtime, domain.InstanceId)
		if errQueryDnsHost != nil {
			loger.Debug("errQueryDnsHost: %v", errQueryDnsHost)
		}
		for _, dhl := range respQuery.Body.DnsHostList {
			for _, il := range dhl.IpList {
				myip.CurrentCache.Put(cacheKey, *il, ipType, *dhl.DnsName)
			}
		}
		// 有变化： 并且缓存中没有有记录
		if myip.CurrentCache.IsNotExist(cacheKey) {
			// AddDomainRecordRequest := &client.AddDomainRecordRequest{
			// 	DomainName: &Domain,
			// 	RR:         &DomainRR,
			// 	Type:       &ipType,
			// 	Value:      &IP,
			// 	TTL:        &ttl,
			// }

			respCreatingDnsHost, errrespCreatingDnsHost := CreatingDnsHost(cli, runtime, domain.InstanceId, &DnsName, &IP)
			if errrespCreatingDnsHost != nil {
				if ErrorDomainRecordDuplicate(errrespCreatingDnsHost) {
					if respCreatingDnsHost != nil && respCreatingDnsHost.Body != nil {
						// 更新缓存
						myip.CurrentCache.Put(cacheKey, IP, ipType, DnsName)
						loger.Info("添加：已经有记录，更新缓存并跳过添加过程")
					}
					return nil
				}
				return errrespCreatingDnsHost
			}
			respSynchronizingDnsHost, errSynchronizingDnsHost := SynchronizingDnsHost(cli, runtime, domain.InstanceId)
			if errSynchronizingDnsHost != nil {
				loger.Debug("respSynchronizingDnsHost.StatusCode: %v", respSynchronizingDnsHost.StatusCode)
				loger.Debug("errSynchronizingDnsHost: %v", errSynchronizingDnsHost)
			}

			if respCreatingDnsHost != nil && respCreatingDnsHost.Body != nil {
				// 更新缓存
				myip.CurrentCache.Put(cacheKey, IP, ipType, DnsName)
				loger.Info("添加：[%s] 公网IP已添加: %s", cacheKey, IP)
			}
		} else {
			// 有记录，更新
			taskNo := myip.CurrentCache.GetRecordId(cacheKey)
			loger.Info("[%s]", taskNo)
			// updateDomainRecordRequest := &client.UpdateDomainRecordRequest{
			// 	RecordId: &recordId,
			// 	RR:       &DomainRR,
			// 	Type:     &ipType,
			// 	Value:    &IP,
			// }
			respModifyingDnsHost, errModifyingDnsHost := ModifyingDnsHost(cli, runtime, domain.InstanceId, &DnsName, &IP)
			if errModifyingDnsHost != nil {
				if ErrorDomainRecordDuplicate(errModifyingDnsHost) {
					if respModifyingDnsHost != nil && respModifyingDnsHost.Body != nil {
						// 更新缓存
						myip.CurrentCache.Put(cacheKey, IP, ipType, DnsName)
						loger.Info("更新：已经有相同记录，稍后重试")
					}
					return nil
				}
				return errModifyingDnsHost
			}
			respSynchronizingDnsHost, errSynchronizingDnsHost := SynchronizingDnsHost(cli, runtime, domain.InstanceId)
			if errSynchronizingDnsHost != nil {
				loger.Debug("respSynchronizingDnsHost.StatusCode: %v", respSynchronizingDnsHost.StatusCode)
				loger.Debug("errSynchronizingDnsHost: %v", errSynchronizingDnsHost)
			}
			if respModifyingDnsHost != nil && respModifyingDnsHost.Body != nil {
				// 更新缓存
				myip.CurrentCache.Put(cacheKey, IP, ipType, DnsName)
				loger.Info("[%s] 公网IP更新: %s", cacheKey, IP)
			}
		}
	}
	return nil
}

// 分页查询自己账户下的域名列表
func QueryDomain(cli client.Client, runtime *util.RuntimeOptions) (_result *client.QueryDomainListResponse, _err error) {
	queryDomainListRequest := &client.QueryDomainListRequest{
		PageNum:  tea.Int32(1),
		PageSize: tea.Int32(100),
	}
	resp, err := cli.QueryDomainListWithOptions(queryDomainListRequest, runtime)
	return resp, err
}

// 查询域名DNS Host
func QueryDnsHost(cli client.Client, runtime *util.RuntimeOptions, instanceId *string) (_result *client.QueryDnsHostResponse, _err error) {
	queryDnsHostRequest := &client.QueryDnsHostRequest{
		InstanceId: instanceId,
	}
	resp, err := cli.QueryDnsHostWithOptions(queryDnsHostRequest, runtime)

	return resp, err
}

// 提交同步DNS host任务
func SynchronizingDnsHost(cli client.Client, runtime *util.RuntimeOptions, instanceId *string) (_result *client.SaveSingleTaskForSynchronizingDnsHostResponse, _err error) {
	saveSingleTaskForSynchronizingDnsHostRequest := &client.SaveSingleTaskForSynchronizingDnsHostRequest{
		InstanceId: instanceId,
	}
	resp, err := cli.SaveSingleTaskForSynchronizingDnsHostWithOptions(saveSingleTaskForSynchronizingDnsHostRequest, runtime)
	return resp, err
}

// 提交单个创建DNS host任务
func CreatingDnsHost(cli client.Client, runtime *util.RuntimeOptions, instanceId, dnsName, ip *string) (_result *client.SaveSingleTaskForCreatingDnsHostResponse, _err error) {
	saveSingleTaskForCreatingDnsHostRequest := &client.SaveSingleTaskForCreatingDnsHostRequest{
		InstanceId: instanceId,
		DnsName:    dnsName,
		Ip:         []*string{ip},
	}

	resp, err := cli.SaveSingleTaskForCreatingDnsHostWithOptions(saveSingleTaskForCreatingDnsHostRequest, runtime)
	return resp, err
}

// 提交修改DNS host任务
func ModifyingDnsHost(cli client.Client, runtime *util.RuntimeOptions, instanceId, dnsName, ip *string) (_result *client.SaveSingleTaskForModifyingDnsHostResponse, _err error) {
	saveSingleTaskForModifyingDnsHostRequest := &client.SaveSingleTaskForModifyingDnsHostRequest{
		InstanceId: instanceId,
		DnsName:    dnsName,
		Ip:         []*string{ip},
	}
	resp, err := cli.SaveSingleTaskForModifyingDnsHostWithOptions(saveSingleTaskForModifyingDnsHostRequest, runtime)
	return resp, err
}

// 提交删除DNSHost任务
func DeletingDnsHost(cli client.Client, runtime *util.RuntimeOptions, instanceId, dnsName *string) (_result *client.SaveSingleTaskForDeletingDnsHostResponse, _err error) {
	saveSingleTaskForDeletingDnsHostRequest := &client.SaveSingleTaskForDeletingDnsHostRequest{
		InstanceId: instanceId,
		DnsName:    dnsName,
	}
	resp, err := cli.SaveSingleTaskForDeletingDnsHostWithOptions(saveSingleTaskForDeletingDnsHostRequest, runtime)
	return resp, err
}

// 错误处理
func ErrorDomainRecordDuplicate(err error) bool {
	return strings.Contains(err.Error(), "DomainRecordDuplicate")
}
