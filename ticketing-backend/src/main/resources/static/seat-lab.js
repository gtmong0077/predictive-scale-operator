const state = {
    autoRefreshTimer: null,
    seats: []
};

const elements = {
    eventId: document.getElementById('eventId'),
    userId: document.getElementById('userId'),
    loadSeatsButton: document.getElementById('loadSeatsButton'),
    autoRefresh: document.getElementById('autoRefresh'),
    statusBanner: document.getElementById('statusBanner'),
    seatGrid: document.getElementById('seatGrid'),
    actionLog: document.getElementById('actionLog')
};

function getEventId() {
    return Number(elements.eventId.value || 1);
}

function getUserId() {
    return Number(elements.userId.value || 3001);
}

function setBanner(message, tone = 'info') {
    elements.statusBanner.className = `status-banner ${tone}`;
    elements.statusBanner.textContent = message;
}

function addLog(title, detail) {
    const item = document.createElement('article');
    item.className = 'log-entry';

    const time = new Date().toLocaleTimeString('ko-KR', { hour12: false });
    item.innerHTML = `<strong>${title}</strong><span>${time}</span><p>${detail}</p>`;

    elements.actionLog.prepend(item);

    while (elements.actionLog.children.length > 8) {
        elements.actionLog.removeChild(elements.actionLog.lastChild);
    }
}

async function requestJson(url, options = {}) {
    const response = await fetch(url, {
        headers: {
            'Content-Type': 'application/json',
            ...(options.headers || {})
        },
        ...options
    });

    const text = await response.text();
    const body = text ? JSON.parse(text) : null;

    if (!response.ok) {
        const message = body?.message || `HTTP ${response.status}`;
        throw new Error(message);
    }

    return body;
}

function formatSeatMeta(seat) {
    if (seat.status === 'HELD' && seat.heldBy) {
        return `heldBy: ${seat.heldBy}${seat.holdExpiresAt ? ` / expires: ${seat.holdExpiresAt}` : ''}`;
    }

    return '즉시 테스트 가능';
}

function createSeatCard(seat) {
    const card = document.createElement('article');
    card.className = `seat-card ${seat.status.toLowerCase()}`;

    const holdDisabled = seat.status === 'HELD' || seat.status === 'RESERVED';
    const reserveDisabled = seat.status !== 'HELD';

    card.innerHTML = `
        <div class="seat-header">
            <div>
                <strong>${seat.seatNumber}</strong>
                <div class="seat-meta">seatId: ${seat.seatId}</div>
            </div>
            <span class="status-chip ${seat.status.toLowerCase()}">${seat.status}</span>
        </div>
        <div class="seat-meta">${formatSeatMeta(seat)}</div>
        <div class="seat-actions">
            <button class="seat-action hold" type="button" ${holdDisabled ? 'disabled' : ''}>Hold</button>
            <button class="seat-action reserve" type="button" ${reserveDisabled ? 'disabled' : ''}>Reserve</button>
        </div>
    `;

    const holdButton = card.querySelector('.hold');
    const reserveButton = card.querySelector('.reserve');

    holdButton?.addEventListener('click', () => holdSeat(seat));
    reserveButton?.addEventListener('click', () => reserveSeat(seat));

    return card;
}

function renderSeats(seats) {
    elements.seatGrid.innerHTML = '';

    if (seats.length === 0) {
        const empty = document.createElement('div');
        empty.className = 'empty-state';
        empty.textContent = '좌석 데이터가 없습니다.';
        elements.seatGrid.appendChild(empty);
        return;
    }

    seats.forEach((seat) => {
        elements.seatGrid.appendChild(createSeatCard(seat));
    });
}

async function loadSeats() {
    const eventId = getEventId();
    setBanner('좌석 목록을 불러오는 중입니다...', 'info');

    try {
        const seats = await requestJson(`/api/events/${eventId}/seats`);
        state.seats = seats;
        renderSeats(seats);
        setBanner(`이벤트 ${eventId} 좌석 ${seats.length}개를 불러왔습니다.`, 'success');
        addLog('좌석 조회 성공', `이벤트 ${eventId}의 좌석 상태를 새로고침했습니다.`);
    }
    catch (error) {
        renderSeats([]);
        setBanner(error.message, 'error');
        addLog('좌석 조회 실패', error.message);
    }
}

async function holdSeat(seat) {
    const eventId = getEventId();
    const userId = getUserId();
    setBanner(`${seat.seatNumber} 선점을 시도합니다...`, 'info');

    try {
        await requestJson(`/api/events/${eventId}/seats/${seat.seatId}/hold`, {
            method: 'POST',
            body: JSON.stringify({ userId })
        });

        setBanner(`${seat.seatNumber} 좌석을 userId ${userId}로 선점했습니다.`, 'success');
        addLog('Hold 성공', `${seat.seatNumber} 좌석을 userId ${userId}가 선점했습니다.`);
        await loadSeats();
    }
    catch (error) {
        setBanner(error.message, 'error');
        addLog('Hold 실패', error.message);
    }
}

async function reserveSeat(seat) {
    const eventId = getEventId();
    const userId = getUserId();
    setBanner(`${seat.seatNumber} 예약 확정을 시도합니다...`, 'info');

    try {
        await requestJson(`/api/events/${eventId}/seats/${seat.seatId}/reserve`, {
            method: 'POST',
            body: JSON.stringify({ userId })
        });

        setBanner(`${seat.seatNumber} 좌석을 userId ${userId}로 예약 확정했습니다.`, 'success');
        addLog('Reserve 성공', `${seat.seatNumber} 좌석을 userId ${userId}가 예약 확정했습니다.`);
        await loadSeats();
    }
    catch (error) {
        setBanner(error.message, 'error');
        addLog('Reserve 실패', error.message);
    }
}

function syncAutoRefresh() {
    if (state.autoRefreshTimer) {
        clearInterval(state.autoRefreshTimer);
        state.autoRefreshTimer = null;
    }

    if (elements.autoRefresh.checked) {
        state.autoRefreshTimer = window.setInterval(loadSeats, 10000);
        addLog('자동 새로고침 활성화', '10초마다 좌석 상태를 다시 조회합니다.');
    }
}

elements.loadSeatsButton.addEventListener('click', loadSeats);
elements.autoRefresh.addEventListener('change', syncAutoRefresh);

document.addEventListener('DOMContentLoaded', () => {
    setBanner('Seat Lab 준비 완료. 좌석 새로고침을 눌러 시작하세요.', 'info');
    addLog('Seat Lab 시작', '좌석 보드를 불러올 준비가 완료되었습니다.');
    loadSeats();
});