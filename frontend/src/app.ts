import "./router/outlet";

const app = document.getElementById("app");
if (app) {
  app.innerHTML = `<router-outlet></router-outlet>`;
}
