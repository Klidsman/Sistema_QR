import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

const dias = ['Domingo', 'Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado'];

function hoyISO(): string {
  return new Date().toISOString().slice(0, 10);
}

test.describe('Registro de asistencia del docente', () => {
  test('el docente marca a todos presentes y corrige una excepción', async ({ page }) => {
    await ingresarComoAdministrador(page);
    const sufijo = Date.now().toString().slice(-5);
    const curso = `Curso ${sufijo}`;
    const docente = `Docente ${sufijo}`;
    const correoDocente = `docenteclase${sufijo}@colegio.pe`;
    const claveDocente = 'clave-docente-2026';

    await page.goto('/configuracion/cursos');
    await page.getByLabel('Nombre del curso *').fill(curso);
    await page.getByLabel('1.°', { exact: true }).check();
    await page.getByRole('button', { name: 'Agregar curso' }).click();
    await expect(page.getByText('El curso se creó correctamente.')).toBeVisible();

    await page.goto('/configuracion/turnos');
    const horario = page.locator('form[action="/configuracion/turnos/manana"]');
    await horario.getByLabel('Hora de ingreso (Mañana) *').fill('07:30');
    await horario.getByLabel('Tolerancia en minutos (Mañana) *').fill('10');
    await horario.getByLabel('Hora límite de tardanza (Mañana) *').fill('08:00');
    await horario.getByLabel('Hora de cierre (Mañana) *').fill('12:30');
    await horario.getByLabel('Vigencia desde (Mañana) *').fill(hoyISO());
    await horario.getByRole('button', { name: 'Guardar horario' }).click();
    await expect(page.getByText('El horario del turno se guardó correctamente.')).toBeVisible();

    await page.goto('/configuracion/bloques?turno=manana');
    await page.getByLabel('Número de bloque *').fill('1');
    await page.getByLabel('Hora de inicio *').fill('07:45');
    await page.getByLabel('Hora de fin *').fill('08:30');
    await page.getByLabel('Vigencia desde *').fill(hoyISO());
    await page.getByRole('button', { name: 'Agregar bloque' }).click();
    await expect(page.getByText('El bloque de clase se creó correctamente.')).toBeVisible();

    await page.getByLabel('Número de bloque *').fill('2');
    await page.getByLabel('Hora de inicio *').fill('08:30');
    await page.getByLabel('Hora de fin *').fill('09:15');
    await page.getByLabel('Vigencia desde *').fill(hoyISO());
    await page.getByRole('button', { name: 'Agregar bloque' }).click();
    await expect(page.getByText('El bloque de clase se creó correctamente.')).toBeVisible();

    await page.goto('/configuracion/horarios');
    await page.getByLabel('Sección').selectOption({ label: '1.° A' });
    await page.getByLabel('Fecha').fill(hoyISO());
    await page.getByRole('button', { name: 'Ver horario' }).click();
    await page.getByLabel(`${dias[new Date().getDay()]} bloque 1`).selectOption({ label: curso });
    await page.getByLabel(`${dias[new Date().getDay()]} bloque 2`).selectOption({ label: curso });
    await page.getByRole('button', { name: 'Guardar horario' }).click();
    await page.waitForLoadState('load');
    const confirmar = page.getByRole('button', { name: 'Confirmar reemplazo' });
    if ((await confirmar.count()) > 0) {
      await confirmar.click();
    }
    await expect(page.getByText('El horario semanal se guardó correctamente.')).toBeVisible();

    for (const nombre of ['Ana', 'Luis']) {
      await page.goto('/alumnos/nuevo');
      await page.getByLabel('DNI o código *').fill(`9${sufijo}${nombre.length}`);
      await page.getByLabel('Nombres *').fill(nombre);
      await page.getByLabel('Apellidos *').fill(`Clase ${sufijo}`);
      await page.getByLabel('Grado y sección *').selectOption({ label: '1.° A (2026)' });
      await page.getByRole('button', { name: 'Guardar' }).click();
      await expect(page.getByText('El alumno se registró correctamente.')).toBeVisible();
    }

    await page.goto('/docentes');
    await page.getByLabel('Nombre completo *').fill(docente);
    await page.getByLabel('Correo *').fill(correoDocente);
    await page.getByLabel('Contraseña *').fill(claveDocente);
    await page.getByRole('button', { name: 'Crear docente' }).click();
    await expect(page.getByText('La cuenta de docente se creó correctamente.')).toBeVisible();

    await page.goto('/asistencia/asignaciones');
    await page.getByLabel('Docente *').selectOption({ label: docente });
    await page.getByLabel('Curso *').selectOption({ label: curso });
    await page.getByLabel('Sección *').selectOption({ label: '1.° A (2026)' });
    await page.getByLabel('Desde *').fill(hoyISO());
    await page.getByRole('button', { name: 'Crear asignación' }).click();
    await expect(page.getByText('La asignación se creó correctamente.')).toBeVisible();

    await page.getByRole('button', { name: 'Salir' }).click();
    await page.getByLabel('Correo').fill(correoDocente);
    await page.getByLabel('Contraseña').fill(claveDocente);
    await page.getByRole('button', { name: 'Ingresar' }).click();
    await expect(page).toHaveURL(/\/asistencia$/);

    await page.goto('/asistencia');
    await page.locator('a', { hasText: 'bloque 1' }).first().click();
    await expect(page.getByRole('heading', { name: 'Asistencia de clase' })).toBeVisible();

    await page.getByRole('button', { name: 'Marcar a todos presentes' }).click();
    await expect(page.getByText('Se marcó como presentes a los alumnos sin registro.')).toBeVisible();
    await expect(page.locator('tr.bg-amber-50').first()).toBeVisible();

    await page.locator('select[name^="estado_"]').first().selectOption('falta');
    await page.getByRole('button', { name: 'Guardar asistencia' }).click();
    await expect(page.getByText('La asistencia de la clase se guardó correctamente.')).toBeVisible();

    await page.goto('/asistencia');
    await page.locator('a', { hasText: 'bloque 2' }).first().click();
    await page.getByRole('button', { name: 'Marcar a todos presentes' }).click();
    await expect(page.getByText('Se marcó como presentes a los alumnos sin registro.')).toBeVisible();
    await page.locator('select[name^="estado_"]').first().selectOption('tardanza');
    await page.getByRole('button', { name: 'Guardar asistencia' }).click();
    await expect(page.getByText('La asistencia de la clase se guardó correctamente.')).toBeVisible();

    await page.goto('/asistencia');
    await page.locator('a', { hasText: 'bloque 1' }).first().click();
    await expect(page.locator('select[name^="estado_"]').first()).toHaveValue('falta');
  });
});
