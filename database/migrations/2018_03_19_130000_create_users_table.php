<?php

use Illuminate\Support\Facades\Schema;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;

class CreateUsersTable extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('users', function (Blueprint $table) {
            $table->increments('id');
            $table->string('username', 30);
            $table->string('password', 100);
            $table->boolean('blocked')->default(false);
            $table->string('object_type', 10);
            $table->string('back_url', 100)->nullable();
            $table->string('back_key', 100)->nullable();

            $table->string('employer_name', 20)->default('');
            $table->string('employer_surname', 20)->default('');
            $table->string('employer_patronymic', 20)->default('');
            $table->string('company_name', 30)->default('');
            $table->string('phone_number', 15)->default('');

            $table->unique('username');
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::dropIfExists('users');
    }
}
