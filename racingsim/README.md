# Racing sim

Silly "racing sim" to play around with channels. It auto-plays according rules:

- Racing track
  - Fixed length
  - Divided in four segments
  - Each segment has own "terrain type"
- Cars
  - Are of one four types
  - Have +10% or -10% speed on terrain types
  - Can randomly fail (and exit) during the race
  - Can crash into each other (by chance)
- Implementation
  - Single game loop (pretending to be game server)
  - Cars send player control signals (pretending to be online players)
  - UI is done on the CLI
- Race
  - Three laps
  - Four players
