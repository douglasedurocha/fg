package com.example.sistemagestao.controller;

import com.example.sistemagestao.model.Usuario;
import com.example.sistemagestao.service.AuthService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.Data;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/auth")
@RequiredArgsConstructor
@Tag(name = "Autenticação", description = "Endpoints para autenticação de usuários")
public class AuthController {

    private final AuthService authService;

    @PostMapping("/login")
    @Operation(summary = "Login de usuário", description = "Autentica um usuário e retorna um token JWT")
    public ResponseEntity<String> login(@RequestBody LoginRequest request) {
        String token = authService.login(request.getUsername(), request.getPassword());
        return ResponseEntity.ok(token);
    }

    @PostMapping("/registrar")
    @Operation(summary = "Registrar novo usuário", description = "Cria um novo usuário no sistema")
    public ResponseEntity<Usuario> registrar(@RequestBody RegisterRequest request) {
        Usuario usuario = authService.registrar(
            request.getUsername(),
            request.getPassword(),
            request.getNome(),
            request.getEmail()
        );
        return ResponseEntity.ok(usuario);
    }

    @GetMapping("/me")
    @Operation(summary = "Obter usuário atual", description = "Retorna os dados do usuário autenticado")
    public ResponseEntity<Usuario> getCurrentUser() {
        return ResponseEntity.ok(authService.getCurrentUser());
    }
}

@Data
class LoginRequest {
    private String username;
    private String password;
}

@Data
class RegisterRequest {
    private String username;
    private String password;
    private String nome;
    private String email;
} 