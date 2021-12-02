Code for _Public Verification for Private Hash Matching: Challenges, Policy Responses, and Protocols_

# Proof of External Approval of the Hash Set

Generate dummy data
```
mkdir data && python3 generate_data.py <N> <tau> <|X|> ./data
```

Run benchmarks
```
go test -timeout 300m -bench BV
go test -bench QuickVerifier
```

# Guaranteed Eventual Detection Notification

Build and run
```
cd ev && bash build.sh && bash run.sh
```

Then switch directories (in the Docker container) and run benchmarks
```
cd emp/emp-ag2pc
./run ./bin/test_benchmark 12345
```

# Proof of Non-Membership in the Hash Set

Run benchmarks
```
go test -bench NM
```

