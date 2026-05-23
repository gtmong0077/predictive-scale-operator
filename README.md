이 문서는 Kubebuilder로 생성된 오퍼레이터(Operator) 프로젝트의 표준 `README.md` 템플릿입니다. 읽기 편하도록 마크다운 서식을 다듬어 자연스러운 한국어로 번역해 드립니다.

---

# predictive-scale-operator

`// TODO(user): 사용 목적 및 용도에 대한 간단한 개요를 추가하세요.`

## 설명 (Description)

`// TODO(user): 프로젝트에 대한 심층적인 설명과 사용 개요를 작성하세요.`

## 시작하기 (Getting Started)

### 사전 요구 사항 (Prerequisites)

* Go 버전 v1.24.6 이상
* Docker 버전 17.03 이상
* kubectl 버전 v1.11.3 이상
* Kubernetes v1.11.3 이상 클러스터에 대한 접근 권한

### 클러스터에 배포하기 (To Deploy on the cluster)

1. **이미지 빌드 및 푸시:** `IMG` 변수에 지정된 레지스트리 위치로 이미지를 빌드하고 푸시합니다.
```bash
make docker-build docker-push IMG=<some-registry>/predictive-scale-operator:tag

```



```
   > **참고:** 이 이미지는 본인이 지정한 개인 레지스트리에 퍼블리시되어야 합니다. 또한 작업 환경에서 해당 이미지를 풀(pull)할 수 있는 접근 권한이 필요합니다. 위 명령어가 작동하지 않는다면 레지스트리에 적절한 권한이 설정되어 있는지 확인하세요.

2. **CRD 설치:** 클러스터에 CRD(Custom Resource Definitions)를 설치합니다.
   ```bash
make install

```

3. **매니저 배포:** `IMG`로 지정된 이미지를 사용하여 클러스터에 매니저(Manager)를 배포합니다.
```bash
make deploy IMG=<some-registry>/predictive-scale-operator:tag

```



```
   > **참고:** RBAC 에러가 발생할 경우, 클러스터 관리자(`cluster-admin`) 권한을 부여받거나 관리자 계정으로 로그인해야 할 수 있습니다.

4. **솔루션 인스턴스 생성:** `config/samples` 경로에 있는 샘플(예제)을 적용하여 인스턴스를 생성할 수 있습니다.
   ```bash
   kubectl apply -k config/samples/

```

> **참고:** 테스트를 진행하기 전에 샘플 파일에 기본값이 올바르게 설정되어 있는지 확인하세요.

### 제거하기 (To Uninstall)

1. 클러스터에서 인스턴스(CR, Custom Resources)를 삭제합니다:
```bash

```



kubectl delete -k config/samples/

```

2. 클러스터에서 API(CRD)를 삭제합니다:
   ```bash
   make uninstall

```

3. 클러스터에서 컨트롤러 배포를 해제합니다:
```bash
make undeploy

```



```

## 프로젝트 배포 (Project Distribution)

이 솔루션을 사용자에게 릴리스하고 제공하기 위한 옵션은 다음과 같습니다.

### 단일 YAML 번들 제공 방식

레지스트리에 빌드되고 퍼블리시된 이미지를 사용하여 인스톨러를 빌드합니다:
```bash
make build-installer IMG=<some-registry>/predictive-scale-operator:tag

```

> **참고:** 위 Make 타겟을 실행하면 `dist` 디렉토리에 `install.yaml` 파일이 생성됩니다. 이 파일에는 의존성 없이 이 프로젝트를 설치하는 데 필요한, Kustomize로 빌드된 모든 리소스가 포함되어 있습니다.

**인스톨러 사용 방법:**
사용자는 단순히 `kubectl apply -f` 명령어를 실행하여 프로젝트를 설치할 수 있습니다. (예시):

```bash
kubectl apply -f https://raw.githubusercontent.com/<org>/predictive-scale-operator/<tag or branch>/dist/install.yaml

```

### Helm Chart 제공 방식

선택 사항인 Helm 플러그인을 사용하여 차트를 빌드합니다:

```bash
kubebuilder edit --plugins=helm/v2-alpha

```

명령어를 실행하면 `dist/chart` 디렉토리에 차트가 생성되며, 사용자는 이를 통해 솔루션을 설치할 수 있습니다.

> **참고:** 프로젝트 코드를 변경한 경우, 위와 동일한 명령어를 실행하여 Helm Chart를 업데이트하고 최신 변경 사항을 동기화해야 합니다. 또한, 웹훅(webhook)을 생성했다면 `--force` 플래그를 추가하여 명령어를 실행해야 하며, 이전에 `dist/chart/values.yaml`이나 `dist/chart/manager/manager.yaml`에 추가했던 사용자 정의 구성 요소가 있다면 수동으로 다시 적용해 주어야 합니다.

## 기여하기 (Contributing)

`// TODO(user): 다른 사람들이 이 프로젝트에 기여하는 방법에 대한 자세한 가이드를 추가하세요.`

> **참고:** 사용 가능한 모든 `make` 타겟에 대한 자세한 정보는 `make help`를 실행하여 확인할 수 있습니다.

더 자세한 정보는 [Kubebuilder 문서](https://book.kubebuilder.io/)를 참고하세요.

## 라이선스 (License)

Copyright 2026.

Apache License, Version 2.0 (이하 "라이선스")에 따라 라이선스가 부여됩니다. 이 라이선스를 준수하지 않는 한 이 파일을 사용할 수 없습니다. 라이선스 사본은 다음에서 얻을 수 있습니다.

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

관련 법률에서 요구하거나 서면으로 합의하지 않는 한, 본 라이선스에 따라 배포되는 소프트웨어는 명시적이든 묵시적이든 **어떠한 종류의 보증이나 조건 없이 "있는 그대로"** 배포됩니다. 라이선스에 따른 권한 및 제한 사항을 규정하는 특정 언어에 대해서는 라이선스를 참조하십시오.
