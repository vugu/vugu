;; nope doesn't work... (yet)

;; credit: https://github.com/jungomi/xxhash-wasm/blob/master/src/xxhash.wat
(module
  (memory (export "mem") 1)

  (global $PRIME32_1 i32 (i32.const 2654435761))
  (global $PRIME32_2 i32 (i32.const 2246822519))
  (global $PRIME32_3 i32 (i32.const 3266489917))
  (global $PRIME32_4 i32 (i32.const 668265263))
  (global $PRIME32_5 i32 (i32.const 374761393))

  (global $PRIME64_1 i64 (i64.const 11400714785074694791))
  (global $PRIME64_2 i64 (i64.const 14029467366897019727))
  (global $PRIME64_3 i64 (i64.const  1609587929392839161))
  (global $PRIME64_4 i64 (i64.const  9650029242287828579))
  (global $PRIME64_5 i64 (i64.const  2870177450012600261))

  (func (export "xxh32") (param $ptr i32) (param $len i32) (param $seed i32) (result i32)
        (local $h32 i32)
        (local $end i32)
        (local $limit i32)
        (local $v1 i32)
        (local $v2 i32)
        (local $v3 i32)
        (local $v4 i32)
        (set_local $end (i32.add (get_local $ptr) (get_local $len)))
        (if
          (i32.ge_u (get_local $len) (i32.const 16))
          (block
            (set_local $limit (i32.sub (get_local $end) (i32.const 16)))
            (set_local $v1 (i32.add (i32.add (get_local $seed) (get_global $PRIME32_1)) (get_global $PRIME32_2)))
            (set_local $v2 (i32.add (get_local $seed) (get_global $PRIME32_2)))
            (set_local $v3 (i32.add (get_local $seed) (i32.const 0)))
            (set_local $v4 (i32.sub (get_local $seed) (get_global $PRIME32_1)))
            ;; For every chunk of 4 words, so 4 * 32bits = 16 bytes
            (loop $4words-loop
                  (set_local $v1 (call $round32 (get_local $v1) (i32.load (get_local $ptr))))
                  (set_local $ptr (i32.add (get_local $ptr) (i32.const 4)))
                  (set_local $v2 (call $round32 (get_local $v2) (i32.load (get_local $ptr))))
                  (set_local $ptr (i32.add (get_local $ptr) (i32.const 4)))
                  (set_local $v3 (call $round32 (get_local $v3) (i32.load (get_local $ptr))))
                  (set_local $ptr (i32.add (get_local $ptr) (i32.const 4)))
                  (set_local $v4 (call $round32 (get_local $v4) (i32.load (get_local $ptr))))
                  (set_local $ptr (i32.add (get_local $ptr) (i32.const 4)))
                  (br_if $4words-loop (i32.le_u (get_local $ptr) (get_local $limit))))
            (set_local $h32 (i32.add
                              (i32.rotl (get_local $v1) (i32.const 1))
                              (i32.add
                                (i32.rotl (get_local $v2) (i32.const 7))
                                (i32.add
                                  (i32.rotl (get_local $v3) (i32.const 12))
                                  (i32.rotl (get_local $v4) (i32.const 18)))))))
          ;; else block, when input is smaller than 16 bytes
          (set_local $h32 (i32.add (get_local $seed) (get_global $PRIME32_5))))
        (set_local $h32 (i32.add (get_local $h32) (get_local $len)))
        ;; For the remaining words not covered above, either 0, 1, 2 or 3
        (block $exit-remaining-words
               (loop $remaining-words-loop
                     (br_if $exit-remaining-words (i32.gt_u (i32.add (get_local $ptr) (i32.const 4)) (get_local $end)))
                     (set_local $h32 (i32.add (get_local $h32) (i32.mul (i32.load (get_local $ptr)) (get_global $PRIME32_3))))
                     (set_local $h32 (i32.mul (i32.rotl (get_local $h32) (i32.const 17)) (get_global $PRIME32_4)))
                     (set_local $ptr (i32.add (get_local $ptr) (i32.const 4)))
                     (br $remaining-words-loop)))
        ;; For the remaining bytes that didn't make a whole word,
        ;; either 0, 1, 2 or 3 bytes, as 4bytes = 32bits = 1 word.
        (block $exit-remaining-bytes
               (loop $remaining-bytes-loop
                     (br_if $exit-remaining-bytes (i32.ge_u (get_local $ptr) (get_local $end)))
                     (set_local $h32 (i32.add (get_local $h32) (i32.mul (i32.load8_u (get_local $ptr)) (get_global $PRIME32_5))))
                     (set_local $h32 (i32.mul (i32.rotl (get_local $h32) (i32.const 11)) (get_global $PRIME32_1)))
                     (set_local $ptr (i32.add (get_local $ptr) (i32.const 1)))
                     (br $remaining-bytes-loop)))
        ;; Finalise
        (set_local $h32 (i32.xor (get_local $h32) (i32.shr_u (get_local $h32) (i32.const 15))))
        (set_local $h32 (i32.mul (get_local $h32) (get_global $PRIME32_2)))
        (set_local $h32 (i32.xor (get_local $h32) (i32.shr_u (get_local $h32) (i32.const 13))))
        (set_local $h32 (i32.mul (get_local $h32) (get_global $PRIME32_3)))
        (set_local $h32 (i32.xor (get_local $h32) (i32.shr_u (get_local $h32) (i32.const 16))))
        (get_local $h32))

  (func $round32 (param $seed i32) (param $value i32) (result i32)
        (set_local $seed (i32.add  (get_local $seed) (i32.mul (get_local $value) (get_global $PRIME32_2))))
        (set_local $seed (i32.rotl (get_local $seed) (i32.const 13)))
        (set_local $seed (i32.mul (get_local $seed) (get_global $PRIME32_1)))
        (get_local $seed))

  ;; This is the actual WebAssembly implementation.
  ;; It cannot be used directly from JavaScript because of the lack of support
  ;; for i64.
  (func $xxh64 (param $ptr i32) (param $len i32) (param $seed i64) (result i64)
        (local $h64 i64)
        (local $end i32)
        (local $limit i32)
        (local $v1 i64)
        (local $v2 i64)
        (local $v3 i64)
        (local $v4 i64)
        (set_local $end (i32.add (get_local $ptr) (get_local $len)))
        (if
          (i32.ge_u (get_local $len) (i32.const 32))
          (block
            (set_local $limit (i32.sub (get_local $end) (i32.const 32)))
            (set_local $v1 (i64.add (i64.add (get_local $seed) (get_global $PRIME64_1)) (get_global $PRIME64_2)))
            (set_local $v2 (i64.add (get_local $seed) (get_global $PRIME64_2)))
            (set_local $v3 (i64.add (get_local $seed) (i64.const 0)))
            (set_local $v4 (i64.sub (get_local $seed) (get_global $PRIME64_1)))
            ;; For every chunk of 4 words, so 4 * 64bits = 32 bytes
            (loop $4words-loop
                  (set_local $v1 (call $round64 (get_local $v1) (i64.load (get_local $ptr))))
                  (set_local $ptr (i32.add (get_local $ptr) (i32.const 8)))
                  (set_local $v2 (call $round64 (get_local $v2) (i64.load (get_local $ptr))))
                  (set_local $ptr (i32.add (get_local $ptr) (i32.const 8)))
                  (set_local $v3 (call $round64 (get_local $v3) (i64.load (get_local $ptr))))
                  (set_local $ptr (i32.add (get_local $ptr) (i32.const 8)))
                  (set_local $v4 (call $round64 (get_local $v4) (i64.load (get_local $ptr))))
                  (set_local $ptr (i32.add (get_local $ptr) (i32.const 8)))
                  (br_if $4words-loop (i32.le_u (get_local $ptr) (get_local $limit))))
            (set_local $h64 (i64.add
                              (i64.rotl (get_local $v1) (i64.const 1))
                              (i64.add
                                (i64.rotl (get_local $v2) (i64.const 7))
                                (i64.add
                                  (i64.rotl (get_local $v3) (i64.const 12))
                                  (i64.rotl (get_local $v4) (i64.const 18))))))
            (set_local $h64 (call $merge-round64 (get_local $h64) (get_local $v1)))
            (set_local $h64 (call $merge-round64 (get_local $h64) (get_local $v2)))
            (set_local $h64 (call $merge-round64 (get_local $h64) (get_local $v3)))
            (set_local $h64 (call $merge-round64 (get_local $h64) (get_local $v4))))
          ;; else block, when input is smaller than 32 bytes
          (set_local $h64 (i64.add (get_local $seed) (get_global $PRIME64_5))))
        (set_local $h64 (i64.add (get_local $h64) (i64.extend_u/i32 (get_local $len))))
        ;; For the remaining words not covered above, either 0, 1, 2 or 3
        (block $exit-remaining-words
               (loop $remaining-words-loop
                     (br_if $exit-remaining-words (i32.gt_u (i32.add (get_local $ptr) (i32.const 8)) (get_local $end)))
                     (set_local $h64 (i64.xor (get_local $h64) (call $round64 (i64.const 0) (i64.load (get_local $ptr)))))
                     (set_local $h64 (i64.add
                                       (i64.mul
                                         (i64.rotl (get_local $h64) (i64.const 27))
                                         (get_global $PRIME64_1))
                                       (get_global $PRIME64_4)))
                     (set_local $ptr (i32.add (get_local $ptr) (i32.const 8)))
                     (br $remaining-words-loop)))
        ;; For the remaining half word. That is when there are more than 32bits
        ;; remaining which didn't make a whole word.
        (if
          (i32.le_u (i32.add (get_local $ptr) (i32.const 4)) (get_local $end))
          (block
            (set_local $h64 (i64.xor (get_local $h64) (i64.mul (i64.load32_u (get_local $ptr)) (get_global $PRIME64_1))))
            (set_local $h64 (i64.add
                              (i64.mul
                                (i64.rotl (get_local $h64) (i64.const 23))
                                (get_global $PRIME64_2))
                              (get_global $PRIME64_3)))
            (set_local $ptr (i32.add (get_local $ptr) (i32.const 4)))))
        ;; For the remaining bytes that didn't make a half a word (32bits),
        ;; either 0, 1, 2 or 3 bytes, as 4bytes = 32bits = 1/2 word.
        (block $exit-remaining-bytes
               (loop $remaining-bytes-loop
                     (br_if $exit-remaining-bytes (i32.ge_u (get_local $ptr) (get_local $end)))
                     (set_local $h64 (i64.xor (get_local $h64) (i64.mul (i64.load8_u (get_local $ptr)) (get_global $PRIME64_5))))
                     (set_local $h64 (i64.mul (i64.rotl (get_local $h64) (i64.const 11)) (get_global $PRIME64_1)))
                     (set_local $ptr (i32.add (get_local $ptr) (i32.const 1)))
                     (br $remaining-bytes-loop)))
        ;; Finalise
        (set_local $h64 (i64.xor (get_local $h64) (i64.shr_u (get_local $h64) (i64.const 33))))
        (set_local $h64 (i64.mul (get_local $h64) (get_global $PRIME64_2)))
        (set_local $h64 (i64.xor (get_local $h64) (i64.shr_u (get_local $h64) (i64.const 29))))
        (set_local $h64 (i64.mul (get_local $h64) (get_global $PRIME64_3)))
        (set_local $h64 (i64.xor (get_local $h64) (i64.shr_u (get_local $h64) (i64.const 32))))
        (get_local $h64))

  (func $round64 (param $acc i64) (param $value i64) (result i64)
        (set_local $acc (i64.add  (get_local $acc) (i64.mul (get_local $value) (get_global $PRIME64_2))))
        (set_local $acc (i64.rotl (get_local $acc) (i64.const 31)))
        (set_local $acc (i64.mul (get_local $acc) (get_global $PRIME64_1)))
        (get_local $acc))

  (func $merge-round64 (param $acc i64) (param $value i64) (result i64)
        (set_local $value (call $round64 (i64.const 0) (get_local $value)))
        (set_local $acc (i64.xor (get_local $acc) (get_local $value)))
        (set_local $acc (i64.add (i64.mul (get_local $acc) (get_global $PRIME64_1)) (get_global $PRIME64_4)))
        (get_local $acc))

  ;; This function can be called from JavaScript and it expects that the
  ;; first word in the memory is the u64 seed, which is followed by the actual
  ;; data that is being hashed.
  ;; $ptr indicates the beginning of the memory where it's stored (with seed).
  ;; $len is the length of the actual data (without the 8bytes for the seed).
  ;; The function itself doesn't return anything, since the u64 wouldn't work
  ;; in JavaScript, so instead it is stored in place of the seed.
  (func (export "xxh64") (param $ptr i32) (param $len i32)
        (local $seed i64)
        (local $initial-ptr i32)
        (local $h64 i64)
        (set_local $initial-ptr (i32.add (get_local $ptr) (i32.const 0)))
        ;; Assemble the u64 seed from two u32 that were stored from JavaScript.
        ;; I would have thought it would be okay to just load an i64 directly,
        ;; but apparently that is not the case.
        (set_local $seed (i64.or
                           (i64.shl
                             (i64.load32_u (get_local $ptr))
                             (i64.const 32))
                           (i64.load32_u (i32.add (get_local $ptr) (i32.const 4)))))
        (set_local $ptr (i32.add (get_local $ptr) (i32.const 8)))
        (set_local $h64 (call $xxh64 (get_local $ptr) (get_local $len) (get_local $seed)))
        ;; Disassemble the u64 hash result to two u32 that can be read from
        ;; JavaScript. Again, I would have thought just storing the i64 would be
        ;; good enough.
        (i32.store (get_local $initial-ptr) (i32.wrap/i64 (i64.shr_u (get_local $h64) (i64.const 32))))
        (i32.store (i32.add (get_local $initial-ptr) (i32.const 4)) (i32.wrap/i64 (get_local $h64)))))
