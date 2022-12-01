Code for _Public Verification for Private Hash Matching: Challenges, Policy Responses, and Protocols_

# Proof of External Approval of the Hash Set

Generate dummy data
```
cd bv && mkdir data && python3 generate_data.py <N> <tau> <|X|> ./data
```

Run benchmarks
```
go test -timeout 300m -bench BV
go test -bench QuickVerifier
```

# Guaranteed Eventual Detection Notification

Requires [`docker`](https://www.docker.com)

Build and run `docker` container. 
```
cd ev && bash build.sh && bash run.sh
```

Then switch directories (in the `docker` container) and run benchmarks
```
cd emp/emp-ag2pc
./run ./bin/test_benchmark 12345
```

# Proof of Non-Membership in the Hash Set

Run benchmarks (both interactive and non-interactive proofs) 
```
cd nm && go test -bench NM
```

Run benchmarks (only non-interactive proof) 
```
gcd nm && go test -bench NMFS
```