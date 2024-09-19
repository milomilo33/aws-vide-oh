#![feature(decl_macro, proc_macro_hygiene)]
#[macro_use]
extern crate diesel;
extern crate dotenv;
#[macro_use]
extern crate rocket;
#[macro_use]
extern crate serde_derive;

use dotenv::dotenv;
use lambda_web::{is_running_on_lambda, launch_rocket_on_lambda, LambdaError};
use rocket::tokio;

mod schema;
mod connection;
mod models;
mod repository;
mod handler;
mod router;
mod auth;
mod cors;

// use diesel_migrations::{embed_migrations, EmbeddedMigrations, MigrationHarness};

use diesel::prelude::*;
use crate::connection::{get_connection_string, init_pool};
use diesel_async_migrations::EmbeddedMigrations;
use diesel_async::{AsyncPgConnection, RunQueryDsl, AsyncConnection};

use anyhow::Error;
use std::sync::Arc;

pub const MIGRATIONS: diesel_async_migrations::EmbeddedMigrations = diesel_async_migrations::embed_migrations!();

// async fn run_migrations(pool: Pool<AsyncDieselConnectionManager<AsyncPgConnection>>) -> anyhow::Result<()> {
//     // let pool = Arc::new(init_pool().await);
//     // Introduce a scope to limit the lifetime of the connection
//     // {
//     //     let mut conn = pool.get().await?;
//     //     MIGRATIONS.run_pending_migrations(&mut conn).await?;
//     // }

//     tokio::spawn(async move {
//         let mut conn = pool.get().await.unwrap();
//         MIGRATIONS.run_pending_migrations(&mut conn).await.unwrap();
//         println!("Migrations applied successfully");
//     }).await??;

//     println!("Migrations applied successfully");
//     Ok(())
// }

// async fn run_migrations(pool: Pool<AsyncDieselConnectionManager<AsyncPgConnection>>) -> anyhow::Result<()> {
//     tokio::spawn(async move {
//         let mut conn = match pool.get().await {
//             Ok(conn) => conn,
//             Err(e) => {
//                 eprintln!("Failed to get connection: {:?}", e);
//                 return;
//             }
//         };

//         if let Err(e) = MIGRATIONS.run_pending_migrations(&mut conn).await {
//             eprintln!("Failed to run migrations: {:?}", e);
//         } else {
//             println!("Migrations applied successfully");
//         }
//     })
//     .await
//     .unwrap(); // Handle JoinHandle

//     Ok(())
// }

async fn run_migrations(connection_string: String) -> anyhow::Result<()> {
    let mut conn = diesel_async::AsyncPgConnection::establish(&connection_string).await?;
    MIGRATIONS.run_pending_migrations(&mut conn).await?;
    Ok(())
}

#[rocket::main]
async fn main() -> Result<(), LambdaError> {
    println!("hiiiiiii!!!!!");
    dotenv().ok();

    // Fetch connection string
    let connection_string = get_connection_string().await.expect("Failed to get DB connection string");

    // Run migrations
    if let Err(e) = run_migrations(connection_string.clone()).await {
        eprintln!("Failed to run migrations: {:?}", e);
        return Err(LambdaError::from(e));
    }

    let rocket = router::create_routes(&connection_string ).await?;
    println!("before lambda");
    if is_running_on_lambda() {
        println!("hiiiiiii!!!!!222");
        // Launch on AWS Lambda
        launch_rocket_on_lambda(rocket).await?;
    } else {
        println!("what the fuck?");
        // Launch locally
        rocket.launch().await?;
    }
    
    Ok(())
}

// use tokio_postgres::NoTls;

// #[tokio::main]
// async fn main() {
//     match tokio_postgres::connect("host=videoh-db-rdspostgresinstancef1143a12-1u9zojmspa86.cpi4mwekqm79.eu-central-1.rds.amazonaws.com port=5432 user=postgres password=8G?I&{yU2;FS31X0[XHkN+}\\$<is23bC dbname=videoh", tokio_postgres::NoTls).await
//     {
//         Ok(_) => println!("Successfully connected to database!"),
//         Err(e) => eprintln!("Failed to connect to the database: {:?}", e),
//     }
// }
