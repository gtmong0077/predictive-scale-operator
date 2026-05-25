# Ticketing Backend

## 이 프로젝트가 무엇인가

이 코드는 Team 5 쿠버네티스 프로젝트의 백엔드 서버입니다.
지금 단계에서는 화면 중심 프론트를 만드는 것이 아니라, 요청을 받으면 JSON으로 응답하는 Spring Boot 서버를 만드는 중입니다.

현재는 아래 기능이 구현되어 있습니다.

- 서버 실행 확인 API
- 좌석 목록 조회 API
- 좌석 선점 API
- 좌석 예약 확정 API
- 선점 만료 정리 처리
- Swagger UI 기반 API 테스트 화면
- 직관적 테스트용 Seat Lab 페이지
- MySQL / Redis / Adminer 도커 환경
- localdb 프로필 시작 시 샘플 이벤트 1개와 좌석 3개 자동 생성

## 현재 좌석 상태 흐름

- `AVAILABLE`
- `HELD`
- `RESERVED`

선점 시간은 현재 `3분`으로 설정되어 있습니다.

## 실행 위치

프로젝트 경로:

`C:\dev\Kubernetes\predictive-scale-operator\ticketing-backend`

## 준비물

- Java 17
- Docker Desktop
- VS Code

## 실행 모드

- `demo`
  - Docker 없이 서버 기본 동작만 확인하는 모드
  - 좌석 조회는 메모리 데이터로 동작
  - 좌석 선점 / 예약 확정 API는 사용할 수 없음
- `localdb`
  - MySQL / Redis와 연결되는 실제 개발 모드
  - 좌석 조회 / 선점 / 예약 확정 / Swagger UI / Seat Lab 테스트를 확인할 때 사용하는 모드

## 실행 방법

### 1. Docker 인프라 실행

```powershell
docker compose up -d
```

### 2. Spring Boot 서버 실행

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\gradle.ps1 bootRun --args=--spring.profiles.active=localdb
```

## 현재 확인 가능한 주소

- `http://localhost:8080/api/hello`
- `http://localhost:8080/api/events/1/seats`
- `http://localhost:8080/actuator/health`
- `http://localhost:8080/swagger-ui.html`
- `http://localhost:8080/seat-lab.html`
- `http://localhost:8080/v3/api-docs`

## 현재 API

### 서버 확인

`GET /api/hello`

### 좌석 목록 조회

`GET /api/events/{eventId}/seats`

예시:

`GET /api/events/1/seats`

### 좌석 선점

`POST /api/events/{eventId}/seats/{seatId}/hold`

요청 본문 예시:

```json
{
  "userId": 3001
}
```

예시:

`POST /api/events/1/seats/1/hold`

### 좌석 예약 확정

`POST /api/events/{eventId}/seats/{seatId}/reserve`

요청 본문 예시:

```json
{
  "userId": 3001
}
```

예시:

`POST /api/events/1/seats/1/reserve`

## 만료 처리 동작

- 만료된 `HELD` 좌석은 조회 시 실제 DB 상태도 `AVAILABLE`로 정리됩니다.
- 만료된 좌석을 바로 예약 확정하려 하면 `409 Conflict`가 반환됩니다.
- 즉, 만료된 좌석은 다시 `hold`부터 잡아야 합니다.

## fresh DB 기준 샘플 데이터

처음 DB가 비어 있을 때만 아래 데이터가 자동으로 들어갑니다.
이미 실행한 뒤에는 데이터가 그대로 남아 있을 수 있습니다.

이벤트 1개:
- `Team 5 Ticketing Demo`

좌석 3개:
- `A-1`: `AVAILABLE`
- `A-2`: `HELD`
- `A-3`: `RESERVED`

## 완전히 새로 시작하고 싶을 때

```powershell
docker compose down -v
docker compose up -d
```

위 명령은 MySQL / Redis 볼륨까지 지우기 때문에 샘플 데이터를 다시 처음 상태로 만들 때 사용합니다.

## Seat Lab으로 테스트하는 방법

브라우저에서 아래 주소로 들어갑니다.

- `http://localhost:8080/seat-lab.html`

이 페이지에서는 아래 흐름을 바로 테스트할 수 있습니다.

- eventId 입력
- userId 입력
- 좌석 목록 새로고침
- 좌석 카드에서 Hold 버튼 클릭
- 같은 좌석을 Reserve 버튼으로 예약 확정
- 10초 자동 새로고침으로 상태 변화 확인

## Swagger UI로 테스트하는 방법

브라우저에서 아래 주소로 들어갑니다.

- `http://localhost:8080/swagger-ui.html`

여기서 바로 아래 API를 눌러 테스트할 수 있습니다.

- `GET /api/hello`
- `GET /api/events/{eventId}/seats`
- `POST /api/events/{eventId}/seats/{seatId}/hold`
- `POST /api/events/{eventId}/seats/{seatId}/reserve`

## PowerShell로 직접 테스트하는 방법

### 1. 좌석 목록 조회

```powershell
Invoke-RestMethod http://localhost:8080/api/events/1/seats | ConvertTo-Json -Depth 6
```

### 2. 비어 있는 좌석 선점

```powershell
Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/api/events/1/seats/1/hold `
  -ContentType 'application/json' `
  -Body '{"userId":3001}' | ConvertTo-Json -Depth 6
```

### 3. 같은 사용자로 예약 확정

```powershell
Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/api/events/1/seats/1/reserve `
  -ContentType 'application/json' `
  -Body '{"userId":3001}' | ConvertTo-Json -Depth 6
```

### 4. 다시 좌석 목록 조회

```powershell
Invoke-RestMethod http://localhost:8080/api/events/1/seats | ConvertTo-Json -Depth 6
```

## 헬스 체크와 메트릭

- Health: `http://localhost:8080/actuator/health`
- Prometheus: `http://localhost:8080/actuator/prometheus`
- OpenAPI JSON: `http://localhost:8080/v3/api-docs`

## MySQL 확인용 Adminer

주소:

- `http://localhost:8081`

접속 정보:

- System: `MySQL`
- Server: `mysql`
- Username: `ticketing`
- Password: `ticketing1234`
- Database: `ticketing`

## 지금까지 구현된 것

1. Spring Boot 프로젝트 뼈대 구성
2. Java 17 고정 실행 스크립트 구성
3. Docker 기반 MySQL / Redis / Adminer 구성
4. `Event`, `Seat` JPA 엔티티 구성
5. 좌석 목록 조회 API 구현
6. Redis 기반 좌석 선점 API 구현
7. Redis 기반 좌석 예약 확정 API 구현
8. 만료된 선점 자동 정리 처리 구현
9. Swagger UI 기반 API 문서 / 테스트 화면 구현
10. Seat Lab 기반 직관적 테스트 페이지 구현

## 다음 단계

다음 구현 목표는 아래 순서가 좋습니다.

1. 백엔드 Dockerfile 추가
2. Kubernetes 배포용 환경 변수 정리
3. k6 테스트를 위한 시나리오 정리
4. 운영용 매니페스트 연결 검증