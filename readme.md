## welcome to drumstick
drumstick is a  locally ran small twitter clone.
The goal of this project is to be able to render a user's 
timeline.

## Prior to Running

### Programs to be installed

This program requires the go programming language
PostgreSQL and a terminal to be installed on your local machine.
Postman is also required to test routes and authentication.

### Installing the Dependencies

#### MacOS
- On Macos: The easiest way to install the above programs is via homebrew pkg manager.
  Open up the terminal and type the cmd `brew install go postgresql`.
  
#### Windows
- On Windows: The easiest way to install the programs is to use the  chocolatey pkg manager.
  Open the terminal and type `choco install go postgresql`. 
  
#### Postman
Postman can be downloaded from the website.

### Setting Up Postgresql

#### Macos
- Open the terminal and start the postgresql server with `brew services start postgresql`

- Included in the project is a sql dump file thats meant to be used to import the database.
  To import the database using the following command. 
  Note: make sure to replace the angular brackets in username with your actual database username.
  `createdb drumstick`.
  `psql --username=<user_name> drumstick < drumstick.sql'`.
 
#### Windows

- Open the services.msc application, look for postgreSQL and make sure its on. 
  if its not, select to turn it on. Afterwards, in cmd.exe write psql

- Included in the project is a sql dump file thats meant to be used to import the database.
  To import the database using the following command. 
  Note: make sure to replace the angular brackets in username with your actual database username.
  `createdb drumstick`.
  `psql --username=<user_name> drumstick < drumstick.sql'`.

### Running the application

After you finsh setting up the database, just write `go run .`
within the directory of the main.go file in the terminal

### Visiting the routes

To visits the routes provide Postman with the right request method 
and address given below. Postman provides a option to render html.


## Routes

### Homepage

The homepage is the welcome page that contains a welcome message and links to the signup and login page.
The route is located at `Get localhost:8080/`.


### SignUp

The signup page is where you fill out a form to create an account. 
Its route method is `GET localhost:8080/signup`. If account creation fails the errors will be listed
in json format. On success the account is created in the database and your logged into the posts page. 

### View 

Once the user fills out the the loginForm and clicks submit 
a post route with that form data `POST localhost:8080/view` 
creates the user's account. If there is any error with any of the fields
the fields and their errors are returned in json format.

### LoginForm

The loginForm is a form page where users login; the request is at `GET` `localhost:8080/loginForm`.
Once submitted, its authenticated at post route `Post` `localhost:8080/view`.


### Posts wip

The Post page using the route `POST localhost:8080/posts` given the login form data
authenticates the user's credentials. The Posts page is the where the users can view
and modify their posts. This page is still a work in progress.
