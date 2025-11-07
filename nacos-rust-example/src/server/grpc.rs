use rand::Rng;
use tokio::sync::broadcast;
use tonic::{Request, Response, Status};
use tonic::transport::Server;
use crate::server::grpc::worker::{TaskReply, TaskRequest};
use crate::server::grpc::worker::worker_service_server::{WorkerService, WorkerServiceServer};

pub mod worker{
    tonic::include_proto!("worker_grpc");
}

#[derive(Default)]
pub struct DefaultWorkerService{}
#[tonic::async_trait]
impl WorkerService for DefaultWorkerService {
    async fn work(&self, req: Request<TaskRequest>) -> Result<Response<TaskReply>, Status> {
        let task_request = req.into_inner();
        let mut task_reply = TaskReply::default();
        let mut rng = rand::rng();
        if rng.random_range(1..=100) > 50 {
            // success
            let payload =  task_request.task;
            task_reply.success = true;
            task_reply.solution = payload;
            println!("任务执行成功: {}", task_request.task_id);
        }else {
            // failed
            task_reply.success = false;
            task_reply.error_code = String::from("ERROR_NO_SLOT_AVAILABLE");
            task_reply.error_msg = "服务器资源不足,请稍后重试".to_string();
            println!("任务执行失败: {}", task_request.task_id);
        }
        Ok(Response::new(task_reply))
    }
}

pub async fn grpc_serve(grpc_port: i32, rx: broadcast::Receiver<()>) -> anyhow::Result<()>
{
    let addr = format!("0.0.0.0:{}", grpc_port).parse()?;
    Server::builder()
        .add_service(WorkerServiceServer::new(DefaultWorkerService::default()))
        .serve_with_shutdown(addr, blocking_waiting_shutdown(rx))
        .await?;
    Ok(())
}

async fn blocking_waiting_shutdown(mut rx: broadcast::Receiver<()>) {
    let _ = rx.recv().await;
    tracing::info!("[gRPC] received signal, shutting down");
}