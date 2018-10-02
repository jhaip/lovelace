void pong(int i);

void ping(int i) {
  pong(i);
}

void pong(int i) {
  if (i > 0) {
    ping(i-1);
  }
}

int main (void)
{
  ping(1000);
  return 0;
}
