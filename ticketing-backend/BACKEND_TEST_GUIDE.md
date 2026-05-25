# Ticketing Backend Test Guide

## 1. 문서 목적

이 문서는 어승효 담당 파트인 `ticketing-backend`를 직접 테스트할 때 필요한 명령어와 확인 방법을 한 번에 정리한 문서입니다.
Windows 11 + VS Code + PowerShell 기준으로 작성했습니다.

## 2. 현재 구현 범위

현재 테스트 가능한 기능은 아래와 같습니다.

- 서버 실행 확인 API
- 좌석 목록 조회 API
- 좌석 선점 API
- 좌석 예약 확정 API
- 만료된 선점 자동 정리 처리
- Swagger UI 테스트 화면
- Seat Lab 테스트 페이지
- Health / Prometheus / Adminer 확인
- Docker 이미지 빌드 및 컨테이너 실행

## 3. 작업 위치

프로젝트 루트:

`C:\dev\Kubernetes\predictive-scale-operator\ticketing-backend`

PowerShell에서 먼저 아래로 이동합니다.

```powershell
cd C:\dev\Kubernetes\predictive-scale-operator\ticketing-backend
```

## 4. 사전 준비물

- Java 17
- Docker Desktop
- VS Code
- 인터넷 연결
  - Gradle 의존성 다운로드
  - Docker 베이스 이미지 다운로드 시 필요

## 5. 가장 기본적인 실행 순서

### 5.1 Docker 인프라 실행

```powershell
docker compose up -d
```

정상 확인:

```powershell
docker compose ps
```

기대 결과:

- `ticketing-mysql`
- `ticketing-redis`
- `ticketing-adminer`

### 5.2 Spring Boot 서버 실행

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\gradle.ps1 bootRun --args=--spring.profiles.active=localdb
```

중요:

- 이 명령은 끝나지 않는 것이 정상입니다.
- 서버가 켜진 동안 해당 터미널은 계속 점유됩니다.
- 서버를 끄려면 `Ctrl + C`를 누릅니다.

### 5.3 새 터미널 열기

서버 터미널은 그대로 두고, 새 PowerShell 터미널을 하나 더 열어 아래 테스트를 진행합니다.

## 6. 브라우저로 바로 확인할 주소

브라우저에서 직접 열 수 있는 주소입니다.

- 서버 확인: `http://localhost:8080/api/hello`
- 좌석 조회: `http://localhost:8080/api/events/1/seats`
- Health: `http://localhost:8080/actuator/health`
- Prometheus: `http://localhost:8080/actuator/prometheus`
- Swagger UI: `http://localhost:8080/swagger-ui.html`
- Seat Lab: `http://localhost:8080/seat-lab.html`
- OpenAPI JSON: `http://localhost:8080/v3/api-docs`
- Adminer: `http://localhost:8081`

## 7. PowerShell 테스트 명령어

### 7.1 서버 실행 확인

```powershell
Invoke-RestMethod http://localhost:8080/api/hello | ConvertTo-Json -Depth 5
```

정상 기대값 예시:

```json
{
  "message": "Ticketing backend is running",
  "nextStep": "..."
}
```

### 7.2 좌석 목록 조회

```powershell
Invoke-RestMethod http://localhost:8080/api/events/1/seats | ConvertTo-Json -Depth 6
```

정상 기대값 예시:

```json
[
  {
    "seatId": 1,
    "seatNumber": "A-1",
    "status": "AVAILABLE",
    "heldBy": null,
    "holdExpiresAt": null
  }
]
```

### 7.3 비어 있는 좌석 선점 성공 테스트

```powershell
Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/api/events/1/seats/1/hold `
  -ContentType 'application/json' `
  -Body '{"userId":3001}' | ConvertTo-Json -Depth 6
```

성공 시 확인 포인트:

- `status`가 `HELD`
- `heldBy`가 `3001`
- `holdExpiresAt`가 들어 있음

### 7.4 같은 사용자로 예약 확정 성공 테스트

```powershell
Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/api/events/1/seats/1/reserve `
  -ContentType 'application/json' `
  -Body '{"userId":3001}' | ConvertTo-Json -Depth 6
```

성공 시 확인 포인트:

- `status`가 `RESERVED`
- `heldBy`가 `null`
- `holdExpiresAt`가 `null`

### 7.5 선점된 좌석을 다른 사용자가 다시 선점할 때 실패 확인

```powershell
try {
  Invoke-RestMethod -Method Post `
    -Uri http://localhost:8080/api/events/1/seats/1/hold `
    -ContentType 'application/json' `
    -Body '{"userId":4001}' | ConvertTo-Json -Depth 6
} catch {
  $response = $_.Exception.Response
  $reader = New-Object System.IO.StreamReader($response.GetResponseStream())
  $reader.ReadToEnd()
}
```

정상 기대값:

- `409 Conflict`
- `SEAT_NOT_AVAILABLE` 또는 비슷한 충돌 메시지

### 7.6 이미 예약된 좌석 다시 예약 확정 시 실패 확인

```powershell
try {
  Invoke-RestMethod -Method Post `
    -Uri http://localhost:8080/api/events/1/seats/1/reserve `
    -ContentType 'application/json' `
    -Body '{"userId":3001}' | ConvertTo-Json -Depth 6
} catch {
  $response = $_.Exception.Response
  $reader = New-Object System.IO.StreamReader($response.GetResponseStream())
  $reader.ReadToEnd()
}
```

정상 기대값:

- `409 Conflict`
- `SEAT_RESERVE_CONFLICT`

### 7.7 만료된 선점 테스트

1. 먼저 비어 있는 좌석을 선점합니다.

```powershell
Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/api/events/1/seats/2/hold `
  -ContentType 'application/json' `
  -Body '{"userId":3001}' | ConvertTo-Json -Depth 6
```

2. 약 3분 이상 기다립니다.

3. 다시 좌석 목록을 조회합니다.

```powershell
Invoke-RestMethod http://localhost:8080/api/events/1/seats | ConvertTo-Json -Depth 6
```

4. 기대 결과를 확인합니다.

- 해당 좌석이 `AVAILABLE`로 보임
- 만료된 `HELD` 상태가 조회 시 자동 정리됨

5. 만료된 좌석을 바로 reserve 하면 실패하는지 확인합니다.

```powershell
try {
  Invoke-RestMethod -Method Post `
    -Uri http://localhost:8080/api/events/1/seats/2/reserve `
    -ContentType 'application/json' `
    -Body '{"userId":3001}' | ConvertTo-Json -Depth 6
} catch {
  $response = $_.Exception.Response
  $reader = New-Object System.IO.StreamReader($response.GetResponseStream())
  $reader.ReadToEnd()
}
```

정상 기대값:

- `409 Conflict`
- 만료되었다는 메시지

### 7.8 헬스 체크 확인

```powershell
Invoke-RestMethod http://localhost:8080/actuator/health | ConvertTo-Json -Depth 8
```

정상 기대값:

- `status: UP`
- `db: UP`
- `redis: UP`

### 7.9 Prometheus 메트릭 확인

```powershell
Invoke-WebRequest http://localhost:8080/actuator/prometheus -UseBasicParsing | Select-Object -ExpandProperty StatusCode
```

정상 기대값:

- `200`

## 8. Swagger UI 테스트 방법

브라우저 주소:

- `http://localhost:8080/swagger-ui.html`

테스트 순서:

1. 위 주소로 접속
2. `Ticketing API` 섹션 펼치기
3. `GET /api/hello` 실행
4. `GET /api/events/{eventId}/seats` 실행
5. `POST /api/events/{eventId}/seats/{seatId}/hold` 실행
6. `POST /api/events/{eventId}/seats/{seatId}/reserve` 실행

입력 예시:

- `eventId`: `1`
- `seatId`: `1`
- body:

```json
{
  "userId": 3001
}
```

## 9. Seat Lab 테스트 방법

브라우저 주소:

- `http://localhost:8080/seat-lab.html`

테스트 순서:

1. `eventId`에 `1` 입력
2. `userId`에 `3001` 입력
3. `좌석 새로고침` 버튼 클릭
4. `AVAILABLE` 좌석 카드에서 `Hold` 클릭
5. 같은 좌석 카드에서 `Reserve` 클릭
6. 액션 로그와 상태 배너 확인

확인 포인트:

- 좌석 색상과 상태 칩이 바뀌는지
- 실패 시 에러 메시지가 바로 보이는지
- 자동 새로고침 체크 시 10초마다 새로고침되는지

## 10. Adminer로 DB 직접 확인하는 방법

브라우저 주소:

- `http://localhost:8081`

접속 정보:

- System: `MySQL`
- Server: `mysql`
- Username: `ticketing`
- Password: `ticketing1234`
- Database: `ticketing`

로그인 후 예시 SQL:

```sql
SELECT id, seat_number, status, held_by, hold_expires_at, reserved_by, reserved_at
FROM seats
ORDER BY id;
```

확인 포인트:

- 선점 후 `status = HELD`
- 예약 확정 후 `status = RESERVED`
- 만료 정리 후 `status = AVAILABLE`

## 11. 샘플 DB를 완전히 초기화하고 다시 테스트하는 방법

```powershell
docker compose down -v
docker compose up -d
```

그 다음 서버를 다시 실행합니다.

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\gradle.ps1 bootRun --args=--spring.profiles.active=localdb
```

초기 상태 기대값:

- `A-1`: `AVAILABLE`
- `A-2`: `HELD`
- `A-3`: `RESERVED`

## 12. Dockerfile 테스트 명령어

### 12.1 백엔드 이미지 빌드

```powershell
docker build -t gtmong0077/ticketing-backend:0.1.0 .
```

정상 확인:

```powershell
docker images | findstr ticketing-backend
```

### 12.2 demo 프로필로 컨테이너 실행

```powershell
docker run --rm -p 8080:8080 gtmong0077/ticketing-backend:0.1.0
```

확인:

- `http://localhost:8080/api/hello`
- `http://localhost:8080/swagger-ui.html`
- `http://localhost:8080/seat-lab.html`

### 12.3 localdb 프로필로 컨테이너 실행

먼저 MySQL / Redis를 host에서 켜 둡니다.

```powershell
docker compose up -d
```

그 다음 앱 컨테이너를 실행합니다.

```powershell
docker run --rm -p 8080:8080 `
  -e SPRING_PROFILES_ACTIVE=localdb `
  -e DB_HOST=host.docker.internal `
  -e DB_PORT=3306 `
  -e DB_NAME=ticketing `
  -e DB_USER=ticketing `
  -e DB_PASSWORD=ticketing1234 `
  -e REDIS_HOST=host.docker.internal `
  -e REDIS_PORT=6379 `
  gtmong0077/ticketing-backend:0.1.0
```

확인:

- `http://localhost:8080/actuator/health`
- `db: UP`
- `redis: UP`

## 13. 자주 쓰는 종료 / 정리 명령어

### 서버 종료

서버가 켜진 PowerShell 창에서:

```powershell
Ctrl + C
```

### Docker 인프라 종료

```powershell
docker compose down
```

### Docker 인프라 + 볼륨 삭제

```powershell
docker compose down -v
```

## 14. 자주 발생하는 문제

### 8080 포트가 이미 사용 중일 때

- 기존 Spring Boot 서버가 켜져 있을 수 있습니다.
- 이전 `docker run` 컨테이너가 살아 있을 수 있습니다.
- 먼저 기존 프로세스를 종료한 뒤 다시 실행합니다.

### MySQL / Redis 연결 실패

- Docker Desktop이 꺼져 있을 수 있습니다.
- `docker compose up -d`를 안 했을 수 있습니다.
- 컨테이너 상태를 먼저 확인합니다.

```powershell
docker compose ps
```

### Swagger / Seat Lab은 뜨는데 localdb API가 안 될 때

- 현재 서버가 `demo` 프로필로 떠 있을 가능성이 큽니다.
- 반드시 아래 명령으로 다시 띄웁니다.

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\gradle.ps1 bootRun --args=--spring.profiles.active=localdb
```