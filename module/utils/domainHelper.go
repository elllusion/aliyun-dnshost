package utils

import (
	// "strings"

	// "aliyun-dnshost/module/loger"

	"github.com/alibabacloud-go/domain-20180129/v3/client"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
)

type DomainService struct {
	Client  client.Client
	Runtime *util.RuntimeOptions
}

type DNSHostInfo struct {
	DnsHost []dnshost
}

type dnshost struct {
	DnsName string
	IpType  string
	IP      string
}

// 分页查询自己账户下的域名列表
func (domainService *DomainService) QueryDomain() (result *client.QueryDomainListResponse, err error) {

	queryDomainListRequest := &client.QueryDomainListRequest{
		PageNum:  tea.Int32(1),
		PageSize: tea.Int32(100),
	}
	resp, err := domainService.Client.QueryDomainListWithOptions(queryDomainListRequest, domainService.Runtime)
	return resp, err
}

// 查询域名DNS Host
func (domainService *DomainService) QueryDnsHost(instanceId *string) (result *client.QueryDnsHostResponse, err error) {
	queryDnsHostRequest := &client.QueryDnsHostRequest{
		InstanceId: instanceId,
	}
	resp, err := domainService.Client.QueryDnsHostWithOptions(queryDnsHostRequest, domainService.Runtime)

	return resp, err
}

// 提交同步DNS host任务
func (domainService *DomainService) SynchronizingDnsHost(instanceId *string) (result *client.SaveSingleTaskForSynchronizingDnsHostResponse, err error) {
	saveSingleTaskForSynchronizingDnsHostRequest := &client.SaveSingleTaskForSynchronizingDnsHostRequest{
		InstanceId: instanceId,
	}
	resp, err := domainService.Client.SaveSingleTaskForSynchronizingDnsHostWithOptions(saveSingleTaskForSynchronizingDnsHostRequest, domainService.Runtime)
	return resp, err
}

// 提交单个创建DNS host任务
func (domainService *DomainService) CreatingDnsHost(instanceId, dnsName, ip *string) (result *client.SaveSingleTaskForCreatingDnsHostResponse, err error) {
	saveSingleTaskForCreatingDnsHostRequest := &client.SaveSingleTaskForCreatingDnsHostRequest{
		InstanceId: instanceId,
		DnsName:    dnsName,
		Ip:         []*string{ip},
	}

	resp, err := domainService.Client.SaveSingleTaskForCreatingDnsHostWithOptions(saveSingleTaskForCreatingDnsHostRequest, domainService.Runtime)
	return resp, err
}

// 提交修改DNS host任务
func (domainService *DomainService) ModifyingDnsHost(instanceId, dnsName, ip *string) (result *client.SaveSingleTaskForModifyingDnsHostResponse, err error) {
	saveSingleTaskForModifyingDnsHostRequest := &client.SaveSingleTaskForModifyingDnsHostRequest{
		InstanceId: instanceId,
		DnsName:    dnsName,
		Ip:         []*string{ip},
	}
	resp, err := domainService.Client.SaveSingleTaskForModifyingDnsHostWithOptions(saveSingleTaskForModifyingDnsHostRequest, domainService.Runtime)
	return resp, err
}

// 提交删除DNSHost任务
func (domainService *DomainService) DeletingDnsHost(instanceId, dnsName *string) (result *client.SaveSingleTaskForDeletingDnsHostResponse, err error) {
	saveSingleTaskForDeletingDnsHostRequest := &client.SaveSingleTaskForDeletingDnsHostRequest{
		InstanceId: instanceId,
		DnsName:    dnsName,
	}
	resp, err := domainService.Client.SaveSingleTaskForDeletingDnsHostWithOptions(saveSingleTaskForDeletingDnsHostRequest, domainService.Runtime)
	return resp, err
}
