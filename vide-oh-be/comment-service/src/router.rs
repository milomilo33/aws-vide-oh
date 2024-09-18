use rocket::routes;
use rocket::launch;
use rocket::Build;
use rocket::Rocket;
use rocket_lamb::RocketExt;
use crate::connection;
use crate::handler;

pub async fn create_routes() -> Result<Rocket<rocket::Ignite>, rocket::Error> {
    let pool = connection::init_pool().await;

    let rocket = Rocket::build()
        .manage(pool)
        .mount("/dev/api/comments",
            routes![
                handler::show_all_comments_for_video,
                handler::create_comment,
                handler::show_all_reported_comments,
                handler::delete_comment,
                handler::report_comment,
                handler::create_or_update_rating,
                handler::get_rating_for_video,
                handler::get_rating_for_user
            ]
        );

    // Launch the Rocket instance
    rocket.launch().await
}
