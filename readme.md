## Drumstick

- drumstick is a locally ran program that allows you to create a timeline

## Milestones

- [x] Users can signup and create accounts
- [x] Users can login
- [x] Jwt auth is implemented
- [x] Users can create posts
- [ ] Users can reply to each other's posts
- [x] Users can delete their posts.
- [ ] Users can edit their posts.
- [ ] Secure all user post routes with JWT auth.

## Prior to Running

### Programs to be installed

- This program requires the go programming language,
  PostgreSQL and a terminal to be installed on your local machine.
  This program is meant to run on MacOS and Windows.
 

### Setup

- included in the project is a makefile which will setup the project for you.
  
#### MacOS

- Simply run `make macos-setup`and  it will install the hombrew pkg manager. 
  Then, it will install go, and postgresql using homebrew. After Postgresql is installed,
  it'll create the database drumstick and import the sql dump file included in the project.


#### Windows
- wip

### Starting Postgresql

#### Macos
- Open the terminal and start the postgresql server with `brew services start postgresql`
  Then write `psql drumstick` to view the database.
  
#### Windows

- Open the services.msc application, look for postgreSQL and make sure its on. 
  if its not, select to turn it on. Afterwards, in cmd.exe write `psql drumstick`
  to view the database.
  
### Running the application

- After you finsh setting up the database, just write `go run .`
  within the directory of the main.go file in the terminal.
  Then, start by visiting the homepage  `localhost:8080/`

## Routes

### Homepage

The homepage is the welcome page for the site that  links to the signup and login page.
Its route is `GET localhost:8080/`.


### SignUp

The signup page is where you fill out a form to create an account.
Its route method is `GET localhost:8080/signup`. If account creation fails the errors will be listed
in json format. On success the account is created in the database and your logged into the posts page. 

### Login

Where the user logs in; the request is at `GET` `localhost:8080/login`
and autenticates the credentials at `POST` `localhost:8080/login`.
If successful it redirects to the user's post page, other it errors.


### Posts

After login, the user will have access to their Posts page which is their homepage.
The route is located at ` GET localhost:8080/posts`. The Posts page is the where the users 
can view and submit posts. New posts are submitted via `POST localhost:8080`.
Each Post contains the user's username, the date the post was submitted, and 
the post's body. The max size of the post's body is 350 characters. 

Users can also delete their posts by clicking the delete button on any post
or at `POST localhost:8080/posts/:id`.


#### Current Progress

Currently the user can only create new posts to themselves an delete them.
Users cannot reply to each other users. Posts also, can't be edited.
Being able to make replies and edit your posts are wips.

