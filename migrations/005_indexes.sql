CREATE INDEX IF NOT EXISTS idx_students_prodi_angkatan_active ON students (prodi, angkatan) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_courses_semester ON courses (semester);
CREATE INDEX IF NOT EXISTS idx_enrollments_course_id ON enrollments (course_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_student_tahun ON enrollments (student_id, tahun_akademik);
