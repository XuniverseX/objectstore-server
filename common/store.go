package common

// StoreType 存储类型(表示文件存到哪里)
type StoreType int

const (
	_          StoreType = iota
	StoreLocal           // 节点本地
	StoreMinio           // Ceph集群
	StoreOSS             // 阿里OSS
	StoreMix             // 混合(Minio及OSS)
	StoreAll             // 所有类型的存储都存一份数据
)
