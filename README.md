# WordleSolver

The code runs in 3 modes

1. Interactive Mode - You dont know the answer to the puzzle, and the bot helps you figure the answer.
2. Batch Test Mode - You can run your algorithm on a test set of 2000 words to measure the average performance.
3. Single Test Mode - You can run your algorithm against a word provided by you. 

The mode is currently set by the MODE constant defined in main.go

In interactive mode, the feedback from the computer needs to be provided in the form of a 5 letter string, 
where each character represents a color ( X for Grey, Y for Yellow and G for Green )

Hence to provide feedback such as Grey,Grey,Yellow, Green,Grey you need to enter XXYGX. 

To run the code -
  go run main.go
 
