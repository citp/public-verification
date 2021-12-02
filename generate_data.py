import sys, random, os

def random_lowercase(n):
    min_lc = ord(b'a')
    len_lc = 26
    ba = bytearray(os.urandom(n))
    for i, b in enumerate(ba):
        ba[i] = min_lc + b % len_lc 
    return str(ba)

def intersect(Xs, tau):
  cmap = {}
  for i in range(len(Xs)):
    for s in Xs[i]:
      if s not in cmap:
        cmap[s] = 0
      cmap[s] += 1
  cnt = len([s for s in cmap if cmap[s] >= tau])
  return cnt

def export(Xs, dirpath):
  for i in range(len(Xs)):
    fpath = os.path.join(dirpath, f'{i+1}.dat')
    print(fpath)
    with open(fpath, 'w') as oup:
      Xs[i] = sorted(Xs[i])
      for s in Xs[i]:
        oup.write(s + "\n")

def main():
  N, tau, size, dirpath = int(sys.argv[1]), int(sys.argv[2]), int(sys.argv[3]), sys.argv[4]

  print(f"N = {N}, tau = {tau}, |X| ~ {size}")
  print(f"Generating hashes in directory {dirpath}")

  U = [random_lowercase(15) for i in range(size * 2)]
  print(f"Generated |U| = {len(U)}")
  Xs = []
  for i in range(N):
    print(i)
    Xs.append(random.sample(U, size))
  
  print (intersect(Xs, tau))
  export(Xs, dirpath)


if __name__ == "__main__":
  main()