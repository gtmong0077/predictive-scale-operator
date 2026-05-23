package com.team5.ticketing.application;

import java.time.Duration;
import java.util.List;
import java.util.UUID;
import java.util.function.Supplier;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Profile;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.script.DefaultRedisScript;
import org.springframework.stereotype.Service;

import com.team5.ticketing.exception.SeatLockException;

@Service
@Profile("localdb")
public class RedisSeatLockService {

    private static final String RELEASE_LOCK_SCRIPT =
            "if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end";

    private final StringRedisTemplate stringRedisTemplate;
    private final DefaultRedisScript<Long> unlockScript;
    private final long lockDurationSeconds;

    public RedisSeatLockService(
            StringRedisTemplate stringRedisTemplate,
            @Value("${ticketing.lock-duration-seconds:5}") long lockDurationSeconds
    ) {
        this.stringRedisTemplate = stringRedisTemplate;
        this.lockDurationSeconds = lockDurationSeconds;
        this.unlockScript = new DefaultRedisScript<>();
        this.unlockScript.setScriptText(RELEASE_LOCK_SCRIPT);
        this.unlockScript.setResultType(Long.class);
    }

    public <T> T executeWithSeatLock(Long eventId, Long seatId, Supplier<T> action) {
        String lockKey = "lock:event:" + eventId + ":seat:" + seatId;
        String lockValue = UUID.randomUUID().toString();

        Boolean locked = stringRedisTemplate.opsForValue()
                .setIfAbsent(lockKey, lockValue, Duration.ofSeconds(lockDurationSeconds));

        if (!Boolean.TRUE.equals(locked)) {
            throw new SeatLockException(seatId);
        }

        try {
            return action.get();
        }
        finally {
            releaseLock(lockKey, lockValue);
        }
    }

    private void releaseLock(String lockKey, String lockValue) {
        try {
            stringRedisTemplate.execute(unlockScript, List.of(lockKey), lockValue);
        }
        catch (Exception ignored) {
            // Lock cleanup failure should not hide the original business result.
        }
    }
}