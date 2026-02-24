-- +goose Up
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    parent_id INT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_departments_parent
        FOREIGN KEY (parent_id)
        REFERENCES departments(id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_departments_parent_name
    ON departments (parent_id, name);

CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    department_id INT NOT NULL,
    full_name VARCHAR(200) NOT NULL,
    position VARCHAR(200) NOT NULL,
    hired_at DATE NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_employees_department
        FOREIGN KEY (department_id)
        REFERENCES departments(id)
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE employees;
DROP INDEX idx_departments_parent_name;
DROP TABLE departments;