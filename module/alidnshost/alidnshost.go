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
	"context"

	myConfig "aliyun-dnshost/config"
	"aliyun-dnshost/module/loger"
	"aliyun-dnshost/module/myip"

	aliCliSdk "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/domain-20180129/v3/client"
	"github.com/alibabacloud-go/tea/tea"
)

// 创建客户端
func CreateClient(accessKeyId, accessKeySecret *string) (*client.Client, error) {
	cfg := &aliCliSdk.Config{}
	cfg.AccessKeyId = accessKeyId
	cfg.AccessKeySecret = accessKeySecret
	cfg.Endpoint = tea.String(myConfig.AliyunDomain)
	return client.NewClient(cfg)
}

// 正常宽带
func Run(ctx context.Context, cfg *myConfig.Config, cli client.Client) {
	// 查询IP
	PubIPv4, PubIPv6 := myip.PublilcIPs(cfg.IPv4, cfg.IPv6)

	// 更新所有用户配置
	for _, dnshost := range cfg.DNSHost {
		select {
		case <-ctx.Done():
			return
		default:
			if (PubIPv4 != "") && (dnshost.IPv4EN != false) {
				err := UpdateDomains(cli, nil, dnshost.DnsName, PubIPv4, myConfig.IPv4Type)
				if err != nil {
					loger.Error("IPv4 update failed : %s", err.Error())
				}
			}
			if (PubIPv6 != "") && (dnshost.IPv6EN != false) {
				err := UpdateDomains(cli, nil, dnshost.DnsName, PubIPv6, myConfig.IPv6Type)
				if err != nil {
					loger.Error("IPv6 update failed : %s", err.Error())
				}
			}
		}
	}
}

// 宽带多拨或有多条宽带线路
func RunOnMultiBroadband(ctx context.Context, cfg *myConfig.Config, cli client.Client) {
	// 重复获取公网IP
	broadbandIPv4, broadbandIPv6 := myip.MultiBroadbandPublicIPs(cfg.IPv4, cfg.IPv6, cfg.BroadbandRetry)

	// 更新所有用户配置
	for _, dnshost := range cfg.DNSHost {
		select {
		case <-ctx.Done():
			return
		default:
			if broadbandIPv4 != nil && (dnshost.IPv4EN != false) {
				IP := myip.BroadbandIPFisrt(broadbandIPv4)
				err := UpdateDomains(cli, broadbandIPv4, dnshost.DnsName, IP, myConfig.IPv4Type)
				if err != nil {
					loger.Error("update IPv4 failed : %s", err.Error())
				}
			}
			if broadbandIPv6 != nil && (dnshost.IPv6EN != false) {
				IP := myip.BroadbandIPFisrt(broadbandIPv6)
				err := UpdateDomains(cli, broadbandIPv6, dnshost.DnsName, IP, myConfig.IPv6Type)
				if err != nil {
					loger.Error("update IPv6 failed : %s", err.Error())
				}
			}
		}
	}
}
