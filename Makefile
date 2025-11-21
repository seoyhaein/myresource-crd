# code-generator 버전
# 기본은 latest를 사용하고, 필요하면 make tools CODEGEN_VERSION=v0.24.4 같이 덮어쓴다.
CODEGEN_VERSION ?= latest

MODULE := github.com/seoyhaein/myresource-crd
APIS_PKG := $(MODULE)/pkg/apis
GROUP_VERSION := mygroup.example.com/v1alpha1

# 도구 설치, 일단 go/bin 에만 둠.
.PHONY: tools
tools:
	go install k8s.io/code-generator/cmd/deepcopy-gen@$(CODEGEN_VERSION)
	go install k8s.io/code-generator/cmd/client-gen@$(CODEGEN_VERSION)

# deepcopy 코드 생성
.PHONY: deepcopy
deepcopy: tools
	deepcopy-gen \
		--go-header-file=./hack/boilerplate.go.txt \
    	--bounding-dirs=$(APIS_PKG) \
    	--output-file=zz_generated.deepcopy.go \
    	$(APIS_PKG)/$(GROUP_VERSION)

# clientset 코드 생성 (client-gen)
.PHONY: clientset
clientset: tools
	client-gen \
		--clientset-name clientset \
		--input-base $(APIS_PKG) \
		--input $(GROUP_VERSION) \
		--output-dir pkg/clientset \
		--output-pkg $(MODULE)/pkg/clientset \
		--go-header-file ./hack/boilerplate.go.txt

# 둘 다 한 번에
.PHONY: codegen
codegen: deepcopy clientset

# MyResource 컨트롤러 실행
.PHONY: run-controller
run-controller:
	go run ./cmd/myresource-controller

# MyResource 리소스 확인
.PHONY: check-myresource
check-myresource:
	go run ./cmd/list-myresources