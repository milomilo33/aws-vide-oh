// use diesel::prelude::*;
// use diesel::r2d2::ConnectionManager;
use diesel_async::{pooled_connection::AsyncDieselConnectionManager, AsyncPgConnection, RunQueryDsl};
// use r2d2::Pool;
use bb8::Pool;
// use r2d2_diesel::ConnectionManager;
use rocket::{Request, State};
use rocket::http::Status;
use rocket::request::{self, FromRequest};
use rocket::outcome::Outcome;
use serde::Deserialize;
use rusoto_core::Region;
use rusoto_secretsmanager::SecretsManager;
use rusoto_secretsmanager::{SecretsManagerClient, GetSecretValueRequest}; 
use urlencoding::encode;

type AsyncPool = Pool<AsyncDieselConnectionManager<AsyncPgConnection>>;

#[derive(Deserialize, Debug)]
struct DbSecret {
    username: String,
    password: String,
    host: String,
    port: i32,
    dbname: String,
}

pub async fn init_pool() -> Pool<AsyncDieselConnectionManager<AsyncPgConnection>> {
    println!("hello1");
    let db_secret = get_db_secret().await.expect("Failed to fetch DB secret");
    let url_encoded_password = encode(&db_secret.password);
    let connection_string = format!(
        "postgres://{}:{}@{}:{}/{}?sslmode=disable",
        db_secret.username,
        url_encoded_password,
        db_secret.host,
        db_secret.port,
        db_secret.dbname
    );
    println!("Connection string: {}", connection_string);
    let manager = AsyncDieselConnectionManager::<AsyncPgConnection>::new(connection_string);
    println!("hello2");
    Pool::builder().build(manager).await.expect("Failed to create DB pool")
}

async fn get_db_secret() -> Result<DbSecret, Box<dyn std::error::Error>> {
    let secret_name = std::env::var("DB_SECRET_NAME").expect("DB_SECRET_NAME must be set");
    println!("dbsecret1");
    let client = SecretsManagerClient::new(Region::default());
    println!("dbsecret2");
    let request = GetSecretValueRequest {
        secret_id: secret_name,
        ..Default::default()
    };
    println!("dbsecret3");
    let result = client.get_secret_value(request).await?; // Use `.await()` here
    println!("dbsecret4");
    let secret_string = result.secret_string.expect("SecretString is empty");
    println!("dbsecret5");
    let db_secret: DbSecret = serde_json::from_str(&secret_string)?;
    println!("{:?}", db_secret);
    Ok(db_secret)
}

pub struct DbConn<'a>(pub bb8::PooledConnection<'a, AsyncDieselConnectionManager<AsyncPgConnection>>);

#[rocket::async_trait]
impl<'r> FromRequest<'r> for DbConn<'r> {
    type Error = ();

    async fn from_request(request: &'r Request<'_>) -> request::Outcome<Self, Self::Error> {
        println!("hey! i get to guard");
        let pool = match request.guard::<&State<AsyncPool>>().await {
            Outcome::Success(pool) => pool,
            Outcome::Error(e) => {
                eprintln!("Failed to retrieve the database pool: {:?}", e);
                return Outcome::Error((Status::ServiceUnavailable, ()));
            },
            Outcome::Forward(_) => return Outcome::Forward(Status::Continue),
        };
        match pool.get().await {
            Ok(conn) => {
                println!("Successfully acquired a database connection");
                Outcome::Success(DbConn(conn))
            },
            Err(e) => {
                eprintln!("Failed to get a connection from the pool: {:?}", e);
                Outcome::Error((Status::ServiceUnavailable, ()))
            }
        }
    }
}

impl<'a> std::ops::Deref for DbConn<'a> {
    type Target = AsyncPgConnection;

    fn deref(&self) -> &Self::Target {
        &self.0
    }
}

impl<'a> std::ops::DerefMut for DbConn<'a> {
    fn deref_mut(&mut self) -> &mut Self::Target {
        &mut self.0
    }
}