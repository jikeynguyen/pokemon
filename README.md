Sure, here is the content formatted as a README file:

```markdown
# Pokémon Go Server and Client

This project implements a server and client for a Pokémon catching and battling game. Users can log in, catch Pokémon, view their collection, and battle against each other.

## Prerequisites

1. Go installed on your machine.
2. Properly formatted `users.json` and `pokedex.json` files in the same directory as the server and client code.

## How to Run

### Step 1: Start the Server

1. Open a terminal.
2. Navigate to the directory containing `server.go`.
3. Run the server using the command:
   ```sh
   go run server.go
   ```
4. Ensure the server starts successfully and listens on the specified port (8080).

### Step 2: Start a Client and Log In

1. Open another terminal.
2. Navigate to the directory containing `client.go`.
3. Run the client using the command:
   ```sh
   go run client.go
   ```
4. You will be prompted to enter your username and password. Use the credentials from `users.json`. For example:
   ```plaintext
   Welcome anonymous user!
   Please enter your username and password to continue: Tester1 9uCh4qxBlFqap_-KiqoM68EqO8yYGpKa1c-BCgkOEa4=
   ```

### Step 3: Test Wiki Command

1. After logging in, type `wiki` and follow the prompt to search for a Pokémon or type. For example:
   ```plaintext
   What do you want to do (wiki, list, battle, catch, logout): wiki
   Search wiki with :mons/types + name
   --> mons Dragonite
   ```
2. The server should return the details of the Pokémon Dragonite.

### Step 4: Test Listing User's Pokémon

1. Type `list` and follow the prompt to see your list of Pokémon. For example:
   ```plaintext
   What do you want to do (wiki, list, battle, catch, logout): list
   Do you want to see your list of pokemons or delete one of them? (see / delete): see
   ```

### Step 5: Test Catching Pokémon

1. Type `catch` to start catching Pokémon. You can choose `auto` or `manual` mode. For example:
   ```plaintext
   What do you want to do (wiki, list, battle, catch, logout): catch
   Do you want to catch auto or manual? (auto / manual + direction): auto
   ```
2. The server should provide feedback on the catching process, and caught Pokémon should be added to your list.

### Step 6: Test Battling Another User

1. Ensure you have another client running with a different user logged in (e.g., Tester2).
2. On the first client, type `battle` and provide the opponent's user ID. For example:
   ```plaintext
   What do you want to do (wiki, list, battle, catch, logout): battle
   Enter opponent user ID: 2
   ```
3. Select your Pokémon for battle. You should see a list of your Pokémon and be prompted to choose three:
   ```plaintext
   Choose your 3 Pokémon for battle:
   Your storage:
   0. Charizard
   1. Alakazam
   2. Magneton
   Enter the indices of 3 Pokémon to battle (comma-separated): 0,1,2
   ```
4. The server should simulate the battle, display the result, and provide type effectiveness analysis.

### Step 7: Log Out

1. Type `logout` to log out of the client:
   ```plaintext
   What do you want to do (wiki, list, battle, catch, logout): logout
   Logging out...
   ```
