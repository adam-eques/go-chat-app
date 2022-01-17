import { initializeApp } from "https://www.gstatic.com/firebasejs/9.6.2/firebase-app.js";
import {
  getAuth,
  createUserWithEmailAndPassword,
  signInWithEmailAndPassword,
} from "https://www.gstatic.com/firebasejs/9.6.2/firebase-auth.js";
import { getAnalytics } from "https://www.gstatic.com/firebasejs/9.6.2/firebase-analytics.js";
// TODO: Add SDKs for Firebase products that you want to use
// https://firebase.google.com/docs/web/setup#available-libraries

// Your web app's Firebase configuration
// For Firebase JS SDK v7.20.0 and later, measurementId is optional
const firebaseConfig = {
  apiKey: "AIzaSyB7Hjva6GTelSF0Y6Kk2lQhzW5e3AiM9co",
  authDomain: "chat-app-e2953.firebaseapp.com",
  projectId: "chat-app-e2953",
  storageBucket: "chat-app-e2953.appspot.com",
  messagingSenderId: "758585911599",
  appId: "1:758585911599:web:76354b46d95e40d0ba19da",
  measurementId: "G-D5P1N69S9S",
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);
const analytics = getAnalytics(app);

const auth = getAuth();

function signup(email, password) {
  createUserWithEmailAndPassword(auth, email, password)
    .then((userCredential) => {
      // Signed in
      const user = userCredential.user;
      // ...
    })
    .catch((error) => {
      const errorCode = error.code;
      const errorMessage = error.message;
      console.error(errorCode, errorMessage);
      // ..
    });
}

function login(email, password) {
  return signInWithEmailAndPassword(auth, email, password)
    .then((userCredential) => {
      // Signed in
      const user = userCredential.user;
      return user;
      // ...
    })
    .catch((error) => {
      const errorCode = error.code;
      const errorMessage = error.message;
      console.error(errorCode, errorMessage);
    });
}

const emailInput = document.querySelector("#email");
const passwordInput = document.querySelector("#password");
const tokenInput = document.querySelector("#token");
const submitButton = document.querySelector("#submit");

function handleSubmit(event) {
  if (!tokenInput.value) {
    event.preventDefault();
    const email = emailInput.value;
    const password = passwordInput.value;

    login(email, password).then((user) => {
      if (user) {
        user.getIdToken().then((token) => {
          tokenInput.value = token;
          submitButton.click();
        });
      }
    });
  }
}

document.querySelector("form").addEventListener("submit", handleSubmit);
